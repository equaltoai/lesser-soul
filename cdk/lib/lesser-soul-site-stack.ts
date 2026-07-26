import path from 'node:path';
import { fileURLToPath } from 'node:url';

import {
  CfnOutput,
  Duration,
  RemovalPolicy,
  Stack,
  type StackProps,
  aws_certificatemanager as acm,
  aws_cloudfront as cloudfront,
  aws_cloudfront_origins as origins,
  aws_route53 as route53,
  aws_route53_targets as targets,
  aws_s3 as s3,
  aws_s3_deployment as s3deploy,
} from 'aws-cdk-lib';
import { Construct } from 'constructs';

import { normalizeDeployStage, readWebDomainConfig, route53RecordNameForDomain } from './app-theory-deploy-config.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

function requireDir(name: string, relativePath: string): string {
  const resolved = path.resolve(__dirname, relativePath);
  if (!path.isAbsolute(resolved)) {
    throw new Error(`${name} path must be absolute: ${resolved}`);
  }
  return resolved;
}

function readOptionalContext(scope: Construct, name: string): string | undefined {
  const raw = scope.node.tryGetContext(name);
  if (raw === undefined || raw === null) {
    return undefined;
  }

  const text = String(raw).trim();
  return text.length > 0 ? text : undefined;
}

function rejectLegacyDeployOverrides(scope: Construct): void {
  const contextKeys = ['domainName', 'certificateArn', 'hostedZoneName'] as const;
  const environmentKeys = ['DOMAIN_NAME', 'CERTIFICATE_ARN', 'HOSTED_ZONE_NAME'] as const;
  const suppliedContext = contextKeys.filter((name) => readOptionalContext(scope, name));
  const suppliedEnvironment = environmentKeys.filter((name) => String(process.env[name] ?? '').trim().length > 0);

  if (suppliedContext.length === 0 && suppliedEnvironment.length === 0) {
    return;
  }

  const supplied = [
    ...suppliedContext.map((name) => `CDK context ${name}`),
    ...suppliedEnvironment.map((name) => `environment ${name}`),
  ];
  throw new Error(
    `Legacy deploy overrides are disabled (${supplied.join(', ')}); configure live DNS only through app-theory/app.json lesserSoul.webDomain.live`,
  );
}

export class LesserSoulSiteStack extends Stack {
  constructor(scope: Construct, id: string, props: StackProps = {}) {
    super(scope, id, props);

    const stage = normalizeDeployStage(readOptionalContext(this, 'stage') ?? 'lab');
    rejectLegacyDeployOverrides(this);
    const webDomain = readWebDomainConfig(stage);
    const domainName = webDomain?.domainName;

    const siteBucket = new s3.Bucket(this, 'SiteBucket', {
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      encryption: s3.BucketEncryption.S3_MANAGED,
      enforceSSL: true,
      removalPolicy: RemovalPolicy.DESTROY,
      autoDeleteObjects: true,
    });

    const namespaceBucket = new s3.Bucket(this, 'NamespaceBucket', {
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      encryption: s3.BucketEncryption.S3_MANAGED,
      enforceSSL: true,
      removalPolicy: RemovalPolicy.RETAIN,
    });

    const siteOutputDir = requireDir('site output', '../dist/site');
    new s3deploy.BucketDeployment(this, 'SiteDeployment', {
      sources: [s3deploy.Source.asset(siteOutputDir)],
      destinationBucket: siteBucket,
      prune: true,
      cacheControl: [s3deploy.CacheControl.fromString('public,max-age=0,s-maxage=300,must-revalidate')],
    });

    const namespaceSourceDir = requireDir('namespace assets', '../site/static/ns/agent-attribution');
    new s3deploy.BucketDeployment(this, 'NamespaceDeployment', {
      sources: [s3deploy.Source.asset(namespaceSourceDir)],
      destinationBucket: namespaceBucket,
      destinationKeyPrefix: 'ns/agent-attribution',
      prune: true,
      contentType: 'application/ld+json',
      cacheControl: [s3deploy.CacheControl.fromString('public,max-age=31536000,immutable')],
    });

    const htmlRewrite = new cloudfront.Function(this, 'HtmlRewrite', {
      code: cloudfront.FunctionCode.fromInline(`
function handler(event) {
  var req = event.request;
  var uri = req.uri || '/';

  if (uri === '/ns' || uri.startsWith('/ns/')) return req;

  if (uri.endsWith('/')) {
    req.uri = uri + 'index.html';
    return req;
  }

  var idx = uri.lastIndexOf('/');
  var last = uri.substring(idx + 1);
  if (last.indexOf('.') !== -1) return req;

  req.uri = uri + '/index.html';
  return req;
}
      `.trim()),
    });

    const baseHeadersPolicy = new cloudfront.ResponseHeadersPolicy(this, 'BaseHeadersPolicy', {
      comment: 'Baseline security headers for spec.lessersoul.ai static pages',
      securityHeadersBehavior: {
        strictTransportSecurity: {
          accessControlMaxAge: Duration.days(365 * 2),
          includeSubdomains: true,
          preload: true,
          override: true,
        },
        contentTypeOptions: { override: true },
        frameOptions: { frameOption: cloudfront.HeadersFrameOption.DENY, override: true },
        referrerPolicy: {
          referrerPolicy: cloudfront.HeadersReferrerPolicy.STRICT_ORIGIN_WHEN_CROSS_ORIGIN,
          override: true,
        },
        xssProtection: { protection: true, modeBlock: true, override: true },
      },
    });

    const namespaceHeadersPolicy = new cloudfront.ResponseHeadersPolicy(this, 'NamespaceHeadersPolicy', {
      comment: 'JSON-LD namespace headers for /ns/*',
      corsBehavior: {
        accessControlAllowCredentials: false,
        accessControlAllowHeaders: ['*'],
        accessControlAllowMethods: ['GET', 'HEAD', 'OPTIONS'],
        accessControlAllowOrigins: ['*'],
        originOverride: true,
      },
      securityHeadersBehavior: {
        strictTransportSecurity: {
          accessControlMaxAge: Duration.days(365 * 2),
          includeSubdomains: true,
          preload: true,
          override: true,
        },
        contentTypeOptions: { override: true },
        frameOptions: { frameOption: cloudfront.HeadersFrameOption.DENY, override: true },
        referrerPolicy: {
          referrerPolicy: cloudfront.HeadersReferrerPolicy.STRICT_ORIGIN_WHEN_CROSS_ORIGIN,
          override: true,
        },
        xssProtection: { protection: true, modeBlock: true, override: true },
      },
    });

    const siteOrigin = origins.S3BucketOrigin.withOriginAccessControl(siteBucket);
    const namespaceOrigin = origins.S3BucketOrigin.withOriginAccessControl(namespaceBucket);

    let hostedZone: route53.IHostedZone | undefined;
    let certificate: acm.ICertificate | undefined;

    if (webDomain) {
      hostedZone = route53.HostedZone.fromHostedZoneAttributes(this, 'HostedZone', {
        hostedZoneId: webDomain.hostedZoneId,
        zoneName: webDomain.hostedZoneName,
      });

      certificate = new acm.Certificate(this, 'SiteCertificate', {
        domainName: webDomain.domainName,
        validation: acm.CertificateValidation.fromDns(hostedZone),
        certificateName: `lesser-soul-${stage}-${webDomain.domainName}`,
      });
    }

    const distribution = new cloudfront.Distribution(this, 'Distribution', {
      defaultRootObject: 'index.html',
      domainNames: domainName ? [domainName] : undefined,
      certificate,
      defaultBehavior: {
        origin: siteOrigin,
        allowedMethods: cloudfront.AllowedMethods.ALLOW_GET_HEAD_OPTIONS,
        viewerProtocolPolicy: cloudfront.ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
        cachePolicy: cloudfront.CachePolicy.CACHING_OPTIMIZED,
        responseHeadersPolicy: baseHeadersPolicy,
        functionAssociations: [
          {
            function: htmlRewrite,
            eventType: cloudfront.FunctionEventType.VIEWER_REQUEST,
          },
        ],
      },
      additionalBehaviors: {
        '/ns/*': {
          origin: namespaceOrigin,
          allowedMethods: cloudfront.AllowedMethods.ALLOW_GET_HEAD_OPTIONS,
          viewerProtocolPolicy: cloudfront.ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
          cachePolicy: cloudfront.CachePolicy.CACHING_OPTIMIZED,
          responseHeadersPolicy: namespaceHeadersPolicy,
        },
      },
      errorResponses: [
        {
          httpStatus: 403,
          responseHttpStatus: 404,
          responsePagePath: '/404/index.html',
          ttl: Duration.minutes(1),
        },
        {
          httpStatus: 404,
          responseHttpStatus: 404,
          responsePagePath: '/404/index.html',
          ttl: Duration.minutes(1),
        },
      ],
    });

    if (webDomain && domainName && hostedZone) {
      const recordName = route53RecordNameForDomain(domainName, webDomain.hostedZoneName);

      new route53.ARecord(this, 'AliasRecordA', {
        zone: hostedZone,
        recordName,
        target: route53.RecordTarget.fromAlias(new targets.CloudFrontTarget(distribution)),
      });

      new route53.AaaaRecord(this, 'AliasRecordAaaa', {
        zone: hostedZone,
        recordName,
        target: route53.RecordTarget.fromAlias(new targets.CloudFrontTarget(distribution)),
      });
    }

    const publicHost = domainName ?? distribution.distributionDomainName;

    new CfnOutput(this, 'CloudFrontDomainName', {
      value: distribution.distributionDomainName,
    });

    new CfnOutput(this, 'PublicHost', {
      value: publicHost,
    });

    new CfnOutput(this, 'SiteUrl', {
      value: `https://${publicHost}/`,
    });

    new CfnOutput(this, 'NamespaceUrl', {
      value: `https://${publicHost}/ns/agent-attribution/v1`,
    });

    new CfnOutput(this, 'SiteBucketName', {
      value: siteBucket.bucketName,
    });

    new CfnOutput(this, 'NamespaceBucketName', {
      value: namespaceBucket.bucketName,
    });
  }
}
