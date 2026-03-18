import { App } from 'aws-cdk-lib';

import { LesserSoulSiteStack } from '../lib/lesser-soul-site-stack.js';

const app = new App();
const stage = String(app.node.tryGetContext('stage') ?? 'lab').trim() || 'lab';

new LesserSoulSiteStack(app, `LesserSoulSite-${stage}`, {
  description: `spec.lessersoul.ai static site and namespace (${stage})`,
});
