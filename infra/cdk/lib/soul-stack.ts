import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";
import { SoulStageConfig } from "./stage-config";

export interface SoulStackProps extends cdk.StackProps {
  stage: "lab" | "live";
  stageConfig: SoulStageConfig;
}

export class SoulStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: SoulStackProps) {
    super(scope, id, props);

    new cdk.CfnOutput(this, "Stage", { value: props.stage });
    new cdk.CfnOutput(this, "InstanceDomain", {
      value: props.stageConfig.instanceDomain,
    });

    const ssmBasePath = `/soul/${props.stageConfig.instanceDomain}`;
    new cdk.CfnOutput(this, "SsmBasePath", { value: ssmBasePath });
    new cdk.CfnOutput(this, "InferenceUrlSsmPath", {
      value: `${ssmBasePath}/inference/url`,
    });
    new cdk.CfnOutput(this, "InferenceKeySsmPath", {
      value: `${ssmBasePath}/inference/key`,
    });
    new cdk.CfnOutput(this, "InstanceKeySsmPath", {
      value: `${ssmBasePath}/lesser-host/instance-key`,
    });
  }
}
