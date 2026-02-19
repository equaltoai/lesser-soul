import * as cdk from "aws-cdk-lib";

export type SoulStage = "lab" | "live";

export interface SoulStageConfig {
  instanceDomain: string;
  lesserHostTrustUrl: string;
  soulCreditsPerKTokens: number;
  memoryCuratorSchedule: string;
  tokenRefreshSchedule: string;
  bridgeEnabled: boolean;
  moderatorEnabled: boolean;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function requireString(
  value: Record<string, unknown>,
  key: string,
): string {
  const v = value[key];
  if (typeof v !== "string" || v.trim() === "") {
    throw new Error(`Missing/invalid "${key}" (expected non-empty string).`);
  }
  return v;
}

function requireNumber(value: Record<string, unknown>, key: string): number {
  const v = value[key];
  if (typeof v !== "number" || Number.isNaN(v)) {
    throw new Error(`Missing/invalid "${key}" (expected number).`);
  }
  return v;
}

function requireBoolean(value: Record<string, unknown>, key: string): boolean {
  const v = value[key];
  if (typeof v !== "boolean") {
    throw new Error(`Missing/invalid "${key}" (expected boolean).`);
  }
  return v;
}

export function loadStageConfig(app: cdk.App, stage: SoulStage): SoulStageConfig {
  const raw = app.node.tryGetContext(stage) as unknown;
  if (!isRecord(raw)) {
    throw new Error(
      `Missing/invalid stage config for "${stage}". Expected context key "${stage}" in infra/cdk/cdk.json.`,
    );
  }

  const instanceDomain = requireString(raw, "instanceDomain");
  if (instanceDomain.includes("://")) {
    throw new Error(
      `Invalid "instanceDomain" for "${stage}": expected a domain like "simulacrum.greater.website" (no scheme).`,
    );
  }

  return {
    instanceDomain,
    lesserHostTrustUrl: requireString(raw, "lesserHostTrustUrl"),
    soulCreditsPerKTokens: requireNumber(raw, "soulCreditsPerKTokens"),
    memoryCuratorSchedule: requireString(raw, "memoryCuratorSchedule"),
    tokenRefreshSchedule: requireString(raw, "tokenRefreshSchedule"),
    bridgeEnabled: requireBoolean(raw, "bridgeEnabled"),
    moderatorEnabled: requireBoolean(raw, "moderatorEnabled"),
  };
}

