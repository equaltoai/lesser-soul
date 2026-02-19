import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";

export interface SoulStackProps extends cdk.StackProps {
  stage: "lab" | "live";
}

export class SoulStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: SoulStackProps) {
    super(scope, id, props);

    new cdk.CfnOutput(this, "Stage", { value: props.stage });
  }
}

