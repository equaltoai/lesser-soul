This directory stores reproducible patched vendor artifacts used by the CDK app.

`aws-cdk-lib-2.245.0-brace-expansion-5.0.5.tgz` is based on the upstream `aws-cdk-lib@2.245.0`
package with its bundled `brace-expansion` dependency replaced by `5.0.5` to address
`GHSA-f886-m6hf-6m8v` until AWS ships an official release with the patched bundle.

To rebuild the artifact:

```sh
npm run refresh:aws-cdk-lib-patch
```
