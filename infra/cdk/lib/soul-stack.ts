import * as cdk from "aws-cdk-lib";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as iam from "aws-cdk-lib/aws-iam";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as eventSources from "aws-cdk-lib/aws-lambda-event-sources";
import * as sqs from "aws-cdk-lib/aws-sqs";
import { Construct } from "constructs";
import * as childProcess from "node:child_process";
import * as path from "node:path";
import { SoulStageConfig } from "./stage-config";

export interface SoulStackProps extends cdk.StackProps {
  stage: "lab" | "live";
  stageConfig: SoulStageConfig;
}

export class SoulStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: SoulStackProps) {
    super(scope, id, props);

    cdk.Tags.of(this).add("app", "lesser-soul");
    cdk.Tags.of(this).add("stage", props.stage);
    cdk.Tags.of(this).add("instance-domain", props.stageConfig.instanceDomain);

    new cdk.CfnOutput(this, "Stage", { value: props.stage });
    new cdk.CfnOutput(this, "InstanceDomain", {
      value: props.stageConfig.instanceDomain,
    });

    const ssmBasePath = `/soul/${props.stageConfig.instanceDomain}`;
    new cdk.CfnOutput(this, "SsmBasePath", { value: ssmBasePath });

    const inferenceUrlSsmPath = `${ssmBasePath}/inference/url`;
    const inferenceKeySsmPath = `${ssmBasePath}/inference/key`;
    const instanceKeySsmPath = `${ssmBasePath}/lesser-host/instance-key`;

    new cdk.CfnOutput(this, "InferenceUrlSsmPath", { value: inferenceUrlSsmPath });
    new cdk.CfnOutput(this, "InferenceKeySsmPath", { value: inferenceKeySsmPath });
    new cdk.CfnOutput(this, "InstanceKeySsmPath", { value: instanceKeySsmPath });

    const table = new dynamodb.Table(this, "SoulTable", {
      tableName: `soul-${props.stage}`,
      partitionKey: { name: "pk", type: dynamodb.AttributeType.STRING },
      sortKey: { name: "sk", type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      timeToLiveAttribute: "ttl",
      removalPolicy:
        props.stage === "lab" ? cdk.RemovalPolicy.DESTROY : cdk.RemovalPolicy.RETAIN,
      pointInTimeRecoverySpecification: {
        pointInTimeRecoveryEnabled: props.stage === "live",
      },
    });

    table.addGlobalSecondaryIndex({
      indexName: "instance-tasks",
      partitionKey: { name: "instance_domain", type: dynamodb.AttributeType.STRING },
      sortKey: { name: "created_at", type: dynamodb.AttributeType.STRING },
      projectionType: dynamodb.ProjectionType.ALL,
    });

    table.addGlobalSecondaryIndex({
      indexName: "agent-subtasks",
      partitionKey: { name: "agent_type", type: dynamodb.AttributeType.STRING },
      sortKey: { name: "sk", type: dynamodb.AttributeType.STRING },
      projectionType: dynamodb.ProjectionType.ALL,
    });

    const researcherQueue = new sqs.Queue(this, "ResearcherQueue", {
      queueName: `soul-researcher-${props.stage}`,
      encryption: sqs.QueueEncryption.SQS_MANAGED,
      retentionPeriod: cdk.Duration.days(4),
      visibilityTimeout: cdk.Duration.seconds(60),
    });

    const resultsQueue = new sqs.Queue(this, "ResultsQueue", {
      queueName: `soul-results-${props.stage}`,
      encryption: sqs.QueueEncryption.SQS_MANAGED,
      retentionPeriod: cdk.Duration.days(4),
      visibilityTimeout: cdk.Duration.seconds(60),
    });

    const repoRoot = path.resolve(__dirname, "../../..");
    const goCode = (entry: string): lambda.Code =>
      lambda.Code.fromAsset(repoRoot, {
        exclude: [
          ".git",
          "bin",
          "external",
          "infra/cdk",
          "reference",
        ],
        bundling: {
          image: lambda.Runtime.PROVIDED_AL2023.bundlingImage,
          command: [
            "bash",
            "-c",
            `GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /asset-output/bootstrap ${entry}`,
          ],
          local: {
            tryBundle(outputDir: string) {
              try {
                childProcess.execSync("go version", { stdio: "ignore" });
              } catch {
                return false;
              }

              childProcess.execSync(
                `go build -trimpath -ldflags="-s -w" -o ${path.join(
                  outputDir,
                  "bootstrap",
                )} ${entry}`,
                {
                  cwd: repoRoot,
                  stdio: "inherit",
                  env: {
                    ...process.env,
                    GOOS: "linux",
                    GOARCH: "arm64",
                    CGO_ENABLED: "0",
                  },
                },
              );
              return true;
            },
          },
        },
      });

    const orchestrator = new lambda.Function(this, "Orchestrator", {
      functionName: `soul-orchestrator-${props.stage}`,
      description: "lesser-soul orchestrator (HTTP via Lambda Function URL)",
      runtime: lambda.Runtime.PROVIDED_AL2023,
      architecture: lambda.Architecture.ARM_64,
      handler: "bootstrap",
      code: goCode("./cmd/orchestrator"),
      timeout: cdk.Duration.seconds(30),
      memorySize: 512,
      environment: {
        SOUL_STAGE: props.stage,
        SOUL_INSTANCE_DOMAIN: props.stageConfig.instanceDomain,
        SOUL_STATE_TABLE_NAME: table.tableName,
        SOUL_RESEARCHER_QUEUE_URL: researcherQueue.queueUrl,
        SOUL_RESULTS_QUEUE_URL: resultsQueue.queueUrl,
      },
    });

    const orchestratorUrl = orchestrator.addFunctionUrl({
      authType: lambda.FunctionUrlAuthType.NONE,
      cors: {
        allowedOrigins: ["*"],
        allowedMethods: [lambda.HttpMethod.POST],
        allowedHeaders: ["authorization", "content-type"],
      },
    });

    const agentRunner = new lambda.Function(this, "AgentRunner", {
      functionName: `soul-agent-runner-${props.stage}`,
      description: "lesser-soul agent-runner (SQS)",
      runtime: lambda.Runtime.PROVIDED_AL2023,
      architecture: lambda.Architecture.ARM_64,
      handler: "bootstrap",
      code: goCode("./cmd/agent-runner"),
      timeout: cdk.Duration.seconds(60),
      memorySize: 512,
      environment: {
        SOUL_STAGE: props.stage,
        SOUL_INSTANCE_DOMAIN: props.stageConfig.instanceDomain,
        SOUL_STATE_TABLE_NAME: table.tableName,
        SOUL_RESEARCHER_QUEUE_URL: researcherQueue.queueUrl,
        SOUL_RESULTS_QUEUE_URL: resultsQueue.queueUrl,
      },
    });

    agentRunner.addEventSource(
      new eventSources.SqsEventSource(researcherQueue, { batchSize: 10 }),
    );

    table.grantReadWriteData(orchestrator);
    table.grantReadWriteData(agentRunner);

    researcherQueue.grantSendMessages(orchestrator);
    resultsQueue.grantSendMessages(agentRunner);

    const ssmParamArn = (parameterName: string): string =>
      cdk.Stack.of(this).formatArn({
        service: "ssm",
        resource: "parameter",
        resourceName: parameterName.startsWith("/")
          ? parameterName.slice(1)
          : parameterName,
      });

    const ssmReadPolicy = new iam.PolicyStatement({
      actions: ["ssm:GetParameter", "ssm:GetParameters", "ssm:GetParameterHistory"],
      resources: [
        ssmParamArn(inferenceUrlSsmPath),
        ssmParamArn(inferenceKeySsmPath),
        ssmParamArn(instanceKeySsmPath),
      ],
    });
    orchestrator.addToRolePolicy(ssmReadPolicy);
    agentRunner.addToRolePolicy(ssmReadPolicy);

    new cdk.CfnOutput(this, "SoulTableName", { value: table.tableName });
    new cdk.CfnOutput(this, "ResearcherQueueUrl", {
      value: researcherQueue.queueUrl,
    });
    new cdk.CfnOutput(this, "ResultsQueueUrl", { value: resultsQueue.queueUrl });
    new cdk.CfnOutput(this, "OrchestratorFunctionUrl", {
      value: orchestratorUrl.url,
      exportName: `lesser-soul-${props.stage}-orchestrator-function-url`,
    });
  }
}
