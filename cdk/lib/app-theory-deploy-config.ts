import fs from 'node:fs';
import path from 'node:path';

export type DeployStage = 'lab' | 'live';

export interface WebDomainConfig {
  domainName: string;
  hostedZoneId: string;
  hostedZoneName: string;
}

const appConfigRelativePath = 'app-theory/app.json';
const webDomainConfigPath = 'lesserSoul.webDomain';
const examplePlaceholderPattern = /<[^>]+>|YOUR_[A-Z0-9_]+|\bTODO\b|\bTBD\b|\bPLACEHOLDER\b|\bCHANGEME\b|EXAMPLE/i;

function defaultAppConfigPath(): string {
  return path.resolve(process.cwd(), '..', appConfigRelativePath);
}

function failDeployConfig(stage: string, configPath: string, message: string): never {
  throw new Error(
    `AppTheory deploy config error: ${message}. Expected ${appConfigRelativePath} at ${configPath} to define ${webDomainConfigPath}.${stage}.{domainName,hostedZoneId,hostedZoneName}; live deploys do not accept certificateArn context or CERTIFICATE_ARN environment input.`,
  );
}

function asRecord(value: unknown): Record<string, unknown> | undefined {
  return value && typeof value === 'object' && !Array.isArray(value) ? (value as Record<string, unknown>) : undefined;
}

export function normalizeDeployStage(stage: string): DeployStage {
  const stageKey = stage.trim().toLowerCase();
  if (stageKey !== 'lab' && stageKey !== 'live') {
    throw new Error(`Unsupported deployment stage ${JSON.stringify(stage)}; expected lab or live`);
  }
  return stageKey;
}

function readRequiredDomainField(
  stage: string,
  configPath: string,
  stageConfig: Record<string, unknown>,
  fieldName: 'domainName' | 'hostedZoneName',
): string {
  const raw = stageConfig[fieldName];
  const value = typeof raw === 'string' ? raw.trim().replace(/\.$/, '').toLowerCase() : '';
  if (value === '' || examplePlaceholderPattern.test(value)) {
    failDeployConfig(stage, configPath, `${webDomainConfigPath}.${stage}.${fieldName} is required`);
  }
  const labels = value.split('.');
  if (
    value.length > 253 ||
    labels.length < 2 ||
    labels.some((label) => !/^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/.test(label))
  ) {
    failDeployConfig(stage, configPath, `${webDomainConfigPath}.${stage}.${fieldName} must be a DNS name`);
  }
  return value;
}

function readHostedZoneId(stage: string, configPath: string, stageConfig: Record<string, unknown>): string {
  const raw = stageConfig.hostedZoneId;
  const value = typeof raw === 'string' ? raw.trim() : '';
  if (value === '' || examplePlaceholderPattern.test(value)) {
    failDeployConfig(stage, configPath, `${webDomainConfigPath}.${stage}.hostedZoneId is required`);
  }
  if (!/^Z[A-Z0-9]{4,32}$/.test(value)) {
    failDeployConfig(stage, configPath, `${webDomainConfigPath}.${stage}.hostedZoneId must be an AWS Route53 hosted zone id`);
  }
  return value;
}

export function readWebDomainConfig(stage: string, configPath = defaultAppConfigPath()): WebDomainConfig | undefined {
  const stageKey = normalizeDeployStage(stage);

  let parsed: unknown;
  try {
    parsed = JSON.parse(fs.readFileSync(configPath, 'utf8'));
  } catch (err) {
    const code = (err as NodeJS.ErrnoException).code;
    if (code === 'ENOENT') {
      failDeployConfig(stageKey, configPath, `missing ${appConfigRelativePath}`);
    }
    failDeployConfig(stageKey, configPath, `failed reading ${appConfigRelativePath}: ${String(err)}`);
  }

  const root = asRecord(parsed);
  const lesserSoul = asRecord(root?.lesserSoul);
  const webDomain = asRecord(lesserSoul?.webDomain);
  const stageConfig = asRecord(webDomain?.[stageKey]);

  if (!stageConfig) {
    if (stageKey === 'live') {
      failDeployConfig(stageKey, configPath, `missing ${webDomainConfigPath}.${stageKey} entry`);
    }
    return undefined;
  }

  const domainName = readRequiredDomainField(stageKey, configPath, stageConfig, 'domainName');
  const hostedZoneName = readRequiredDomainField(stageKey, configPath, stageConfig, 'hostedZoneName');
  if (domainName !== hostedZoneName && !domainName.endsWith(`.${hostedZoneName}`)) {
    failDeployConfig(
      stageKey,
      configPath,
      `${webDomainConfigPath}.${stageKey}.domainName must be the hosted-zone apex or a child of hostedZoneName`,
    );
  }

  return {
    domainName,
    hostedZoneId: readHostedZoneId(stageKey, configPath, stageConfig),
    hostedZoneName,
  };
}

export function route53RecordNameForDomain(domainName: string, hostedZoneName: string): string | undefined {
  const normalizedDomain = domainName.trim().replace(/\.$/, '').toLowerCase();
  const normalizedZone = hostedZoneName.trim().replace(/\.$/, '').toLowerCase();
  return normalizedDomain === normalizedZone ? undefined : normalizedDomain;
}
