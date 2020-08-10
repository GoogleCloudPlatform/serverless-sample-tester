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
go build -o sst main.go
```

Authenticate gcloud with your user account:
```bash
gcloud auth login
```

Set defaults for Cloud Run operations, such as setting the region:
```bash
gcloud config set run/region us-central1
gcloud config set run/platform managed
```

## Usage
Run Serverless Sample Tester by passing in the root directory of the sample you wish to test:
```bash
./sst [target-dir]
```

## Build and Deploying Your Sample
This program allows you to specify how you would like it to build and deploy your sample to Cloud Run, fully managed,
in two ways. From highest to lowest, this program uses the following precedence order to find your build and deploy
instructions:

1. README parsing
1. Defaults

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
In the absence of a README, the tool will fall back on reasonable defaults based on whether the sample is Java-based and/or has a Dockerfile.

#### Configuration and Implementation

#### README location
For parsing the README, the tool assumes that it is located in the target directory. If you wish to parse a README file located elsewhere, you can include the README's location
in a `config.yaml` file in the target directory, using the key `readme`. You can specify an absolute directory, or you can simply
specify a directory relative to the sample's directory.

For example, if the README is in the parent directory of the sample:
```text
readme: ../README.md
```

#### Parsing rules
No parsed commands are run through a shell, meaning that the tool will not perform any typical expansions, pipelines, redirections, or other functions. This also means that popular shell builtin commands like `cd`, `export`, `echo`, and
others may not work as expected.

However, any environment variables referenced in the form of `$var` or `${var}` will be expanded. In addition, the tool supports
bash-style multiline commands (non-quoted backslashes at the end of a line that indicate a line continuation).

The Cloud Run region should be set through the `run/region` gcloud property, as described above. Do not set the region through the `--region`
flag in the `gcloud run` commands; the tool may not work as expected.

To accurately splice out the default Cloud Run service name that is used in the README, include this service name
as the environment variable `$CLOUD_RUN_SERVICE_NAME`. If this is not provided, the tool will attempt to parse
what the name is, but it may not always be accurate. For example, if your README contains the command 
```
gcloud run deploy run-mysql --image gcr.io/[YOUR_PROJECT_ID]/run-mysql
```
then `$CLOUD_RUN_SERVICE_NAME` should be set to `run-mysql`.


### Defaults
If comment code tags aren't added to your README, the program will fall back to reasonable defaults to build and deploy
your sample to Cloud Run based on whether your sample is Java-based (has a `pom.xml` in its root directory) or not.
