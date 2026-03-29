#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "$0")/.."

aws_cdk_version="${1:-2.245.0}"
brace_expansion_version="${2:-5.0.5}"

vendor_dir="$PWD/vendor"
tmp_dir="$(mktemp -d)"

cleanup() {
  rm -rf "$tmp_dir"
}

trap cleanup EXIT

pushd "$tmp_dir" >/dev/null

aws_cdk_tgz="$(npm pack "aws-cdk-lib@${aws_cdk_version}" --silent)"
brace_expansion_tgz="$(npm pack "brace-expansion@${brace_expansion_version}" --silent)"

mkdir aws-cdk-lib brace-expansion
tar -xzf "$aws_cdk_tgz" -C aws-cdk-lib
tar -xzf "$brace_expansion_tgz" -C brace-expansion

rm -rf aws-cdk-lib/package/node_modules/brace-expansion
mv brace-expansion/package aws-cdk-lib/package/node_modules/brace-expansion

mkdir -p "$vendor_dir"
out_name="aws-cdk-lib-${aws_cdk_version}-brace-expansion-${brace_expansion_version}.tgz"
tar -czf "$vendor_dir/$out_name" -C aws-cdk-lib package

popd >/dev/null

printf 'Wrote %s\n' "$vendor_dir/$out_name"
