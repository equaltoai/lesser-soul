import { App } from 'aws-cdk-lib';

import { normalizeDeployStage } from '../lib/app-theory-deploy-config.js';
import { LesserSoulSiteStack } from '../lib/lesser-soul-site-stack.js';

const app = new App();
const stage = normalizeDeployStage(String(app.node.tryGetContext('stage') ?? 'lab'));

new LesserSoulSiteStack(app, `LesserSoulSite-${stage}`, {
  description: `spec.lessersoul.ai static site and namespace (${stage})`,
});
