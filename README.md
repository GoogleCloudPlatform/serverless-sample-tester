# Serverless Sample Tester

This tool streamlines the process of testing Google Cloud Platform samples. It is intended to 
identify bugs that arise in the deployment process, and does not cover unit testing. Currently
the tool only supports Cloud Run.

Serverless Sample Tester does the following steps:

1. Deploys the sample to Cloud Run
1. Checks the deployed service for expected responses
1. Returns a log if any tests failed
1. Cleans up created resources

## Getting Started
Build Serverless Sample Tester:
```bash
go build -o sst cmd/main.go
```

Authenticate gcloud with your user account:
```bash
gcloud auth login
```

Consider setting defaults for Cloud Run operations, such as setting the region:
```bash
gcloud config set run/region us-central1
gcloud config set run/platform managed
```

## Usage
Run Serverless Sample Tester by passing in the root directory of the sample you wish to test:
```bash
./sst [target-dir]
```

### README parsing
To parse build and deploy commands from your sample's README, include the following comment code tag before each gcloud command:

```text
[//]: # ({sst-run-unix})
```

For example:
````text
[//]: # ({sst-run-unix})
```
gcloud builds submit --tag=gcr.io/${GOOGLE_CLOUD_PROJECT}/run-mysql
```
````

## Configuration and Implementation

### README location
For parsing the README, the tool assumes that it is located in the target directory. If you wish to parse a README file located elsewhere, you can include the README's location
in a `config.yaml` file in the target directory, using the key `readme`. You can specify an absolute directory, or you can simply
specify a directory relative to the sample's directory.

For example, if the README is in the parent directory of the sample:
```text
readme: ../README.md
```

### Parsing rules
No parsed commands are run through a shell, meaning that the tool will not perform any typical expansions, pipelines, redirections, or other functions. This also means that popular shell builtin commands like `cd`, `export`, `echo`, and
others may not work as expected.

However, any environment variables referenced in the form of `$var` or `${var}` will be expanded. In addition, the tool supports
bash-style multiline commands (non-quoted backslashes at the end of a line that indicate a line continuation).

The Cloud Run region should be set through the `run/region` gcloud property, as described above. Do not set the region through the `--region`
flag in the `gcloud run` commands; the tool may not work as expected.
