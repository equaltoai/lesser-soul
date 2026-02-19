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
    const researcherTokenSsmPath = `${ssmBasePath}/agents/researcher/token`;
    const researcherRefreshSsmPath = `${ssmBasePath}/agents/researcher/refresh`;
    const assistantTokenSsmPath = `${ssmBasePath}/agents/assistant/token`;
    const assistantRefreshSsmPath = `${ssmBasePath}/agents/assistant/refresh`;
    const curatorTokenSsmPath = `${ssmBasePath}/agents/curator/token`;
    const curatorRefreshSsmPath = `${ssmBasePath}/agents/curator/refresh`;
    const customCoderTokenSsmPath = `${ssmBasePath}/agents/custom-coder/token`;
    const customCoderRefreshSsmPath = `${ssmBasePath}/agents/custom-coder/refresh`;
    const customSummarizerTokenSsmPath = `${ssmBasePath}/agents/custom-summarizer/token`;
    const customSummarizerRefreshSsmPath = `${ssmBasePath}/agents/custom-summarizer/refresh`;

    new cdk.CfnOutput(this, "InferenceUrlSsmPath", { value: inferenceUrlSsmPath });
    new cdk.CfnOutput(this, "InferenceKeySsmPath", { value: inferenceKeySsmPath });
    new cdk.CfnOutput(this, "InstanceKeySsmPath", { value: instanceKeySsmPath });
    new cdk.CfnOutput(this, "ResearcherTokenSsmPath", { value: researcherTokenSsmPath });
    new cdk.CfnOutput(this, "ResearcherRefreshSsmPath", { value: researcherRefreshSsmPath });
    new cdk.CfnOutput(this, "AssistantTokenSsmPath", { value: assistantTokenSsmPath });
    new cdk.CfnOutput(this, "AssistantRefreshSsmPath", { value: assistantRefreshSsmPath });
    new cdk.CfnOutput(this, "CuratorTokenSsmPath", { value: curatorTokenSsmPath });
    new cdk.CfnOutput(this, "CuratorRefreshSsmPath", { value: curatorRefreshSsmPath });
    new cdk.CfnOutput(this, "CustomCoderTokenSsmPath", { value: customCoderTokenSsmPath });
    new cdk.CfnOutput(this, "CustomCoderRefreshSsmPath", { value: customCoderRefreshSsmPath });
    new cdk.CfnOutput(this, "CustomSummarizerTokenSsmPath", { value: customSummarizerTokenSsmPath });
    new cdk.CfnOutput(this, "CustomSummarizerRefreshSsmPath", { value: customSummarizerRefreshSsmPath });

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

    const assistantQueue = new sqs.Queue(this, "AssistantQueue", {
      queueName: `soul-assistant-${props.stage}`,
      encryption: sqs.QueueEncryption.SQS_MANAGED,
      retentionPeriod: cdk.Duration.days(4),
      visibilityTimeout: cdk.Duration.seconds(60),
    });

    const curatorQueue = new sqs.Queue(this, "CuratorQueue", {
      queueName: `soul-curator-${props.stage}`,
      encryption: sqs.QueueEncryption.SQS_MANAGED,
      retentionPeriod: cdk.Duration.days(4),
      visibilityTimeout: cdk.Duration.seconds(60),
    });

    const customCoderQueue = new sqs.Queue(this, "CustomCoderQueue", {
      queueName: `soul-custom-coder-${props.stage}`,
      encryption: sqs.QueueEncryption.SQS_MANAGED,
      retentionPeriod: cdk.Duration.days(4),
      visibilityTimeout: cdk.Duration.seconds(60),
    });

    const customSummarizerQueue = new sqs.Queue(this, "CustomSummarizerQueue", {
      queueName: `soul-custom-summarizer-${props.stage}`,
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
        SOUL_ASSISTANT_QUEUE_URL: assistantQueue.queueUrl,
        SOUL_CURATOR_QUEUE_URL: curatorQueue.queueUrl,
        SOUL_CUSTOM_CODER_QUEUE_URL: customCoderQueue.queueUrl,
        SOUL_CUSTOM_SUMMARIZER_QUEUE_URL: customSummarizerQueue.queueUrl,
        SOUL_RESULTS_QUEUE_URL: resultsQueue.queueUrl,
        LESSER_GRAPHQL_URL: `https://${props.stageConfig.instanceDomain}/api/graphql`,
        LESSER_HOST_TRUST_URL: props.stageConfig.lesserHostTrustUrl,
        SOUL_CREDITS_PER_1K_TOKENS: String(props.stageConfig.soulCreditsPerKTokens),
        SOUL_INFERENCE_URL_SSM_PATH: inferenceUrlSsmPath,
        SOUL_INFERENCE_KEY_SSM_PATH: inferenceKeySsmPath,
        SOUL_INSTANCE_KEY_SSM_PATH: instanceKeySsmPath,
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

    if (props.stage === "lab") {
      const mockInference = new lambda.Function(this, "MockInference", {
        functionName: `soul-mock-inference-${props.stage}`,
        description: "lesser-soul mock OpenAI-compatible inference (lab only)",
        runtime: lambda.Runtime.PROVIDED_AL2023,
        architecture: lambda.Architecture.ARM_64,
        handler: "bootstrap",
        code: goCode("./cmd/mock-inference"),
        timeout: cdk.Duration.seconds(10),
        memorySize: 256,
      });

      const mockInferenceUrl = mockInference.addFunctionUrl({
        authType: lambda.FunctionUrlAuthType.NONE,
        cors: {
          allowedOrigins: ["*"],
          allowedMethods: [lambda.HttpMethod.POST],
          allowedHeaders: ["authorization", "content-type"],
        },
      });

      new cdk.CfnOutput(this, "MockInferenceFunctionUrl", {
        value: mockInferenceUrl.url,
      });
    }

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
        SOUL_ASSISTANT_QUEUE_URL: assistantQueue.queueUrl,
        SOUL_CURATOR_QUEUE_URL: curatorQueue.queueUrl,
        SOUL_CUSTOM_CODER_QUEUE_URL: customCoderQueue.queueUrl,
        SOUL_CUSTOM_SUMMARIZER_QUEUE_URL: customSummarizerQueue.queueUrl,
        SOUL_RESULTS_QUEUE_URL: resultsQueue.queueUrl,
        LESSER_GRAPHQL_URL: `https://${props.stageConfig.instanceDomain}/api/graphql`,
        LESSER_HOST_TRUST_URL: props.stageConfig.lesserHostTrustUrl,
        SOUL_CREDITS_PER_1K_TOKENS: String(props.stageConfig.soulCreditsPerKTokens),
        SOUL_INFERENCE_URL_SSM_PATH: inferenceUrlSsmPath,
        SOUL_INFERENCE_KEY_SSM_PATH: inferenceKeySsmPath,
        SOUL_INSTANCE_KEY_SSM_PATH: instanceKeySsmPath,
      },
    });

    orchestrator.addEventSource(
      new eventSources.SqsEventSource(resultsQueue, { batchSize: 10 }),
    );

    agentRunner.addEventSource(
      new eventSources.SqsEventSource(researcherQueue, { batchSize: 10 }),
    );

    agentRunner.addEventSource(
      new eventSources.SqsEventSource(assistantQueue, { batchSize: 10 }),
    );

    agentRunner.addEventSource(
      new eventSources.SqsEventSource(curatorQueue, { batchSize: 10 }),
    );

    agentRunner.addEventSource(
      new eventSources.SqsEventSource(customCoderQueue, { batchSize: 10 }),
    );

    agentRunner.addEventSource(
      new eventSources.SqsEventSource(customSummarizerQueue, { batchSize: 10 }),
    );

    table.grantReadWriteData(orchestrator);
    table.grantReadWriteData(agentRunner);

    researcherQueue.grantSendMessages(orchestrator);
    assistantQueue.grantSendMessages(orchestrator);
    curatorQueue.grantSendMessages(orchestrator);
    customCoderQueue.grantSendMessages(orchestrator);
    customSummarizerQueue.grantSendMessages(orchestrator);
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
        ssmParamArn(researcherTokenSsmPath),
        ssmParamArn(researcherRefreshSsmPath),
        ssmParamArn(assistantTokenSsmPath),
        ssmParamArn(assistantRefreshSsmPath),
        ssmParamArn(curatorTokenSsmPath),
        ssmParamArn(curatorRefreshSsmPath),
        ssmParamArn(customCoderTokenSsmPath),
        ssmParamArn(customCoderRefreshSsmPath),
        ssmParamArn(customSummarizerTokenSsmPath),
        ssmParamArn(customSummarizerRefreshSsmPath),
      ],
    });
    orchestrator.addToRolePolicy(ssmReadPolicy);
    agentRunner.addToRolePolicy(ssmReadPolicy);

    new cdk.CfnOutput(this, "SoulTableName", { value: table.tableName });
    new cdk.CfnOutput(this, "ResearcherQueueUrl", {
      value: researcherQueue.queueUrl,
    });
    new cdk.CfnOutput(this, "AssistantQueueUrl", { value: assistantQueue.queueUrl });
    new cdk.CfnOutput(this, "CuratorQueueUrl", { value: curatorQueue.queueUrl });
    new cdk.CfnOutput(this, "CustomCoderQueueUrl", { value: customCoderQueue.queueUrl });
    new cdk.CfnOutput(this, "CustomSummarizerQueueUrl", { value: customSummarizerQueue.queueUrl });
    new cdk.CfnOutput(this, "ResultsQueueUrl", { value: resultsQueue.queueUrl });
    new cdk.CfnOutput(this, "OrchestratorFunctionUrl", {
      value: orchestratorUrl.url,
      exportName: `lesser-soul-${props.stage}-orchestrator-function-url`,
    });
  }
}
