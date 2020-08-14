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
Run Serverless Sample Tester:
```bash
./sst [sample-dir]
```

### README parsing
To parse build and deploy commands from your README, include the following comment code tag in it:

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

The parsed commands will not be run through a shell, meaning that the program will not perform any expansions,
pipelines, redirections or any other functions that shells are responsible for. This also means that popular shell
builtin commands like `cd`, `export`, and `echo` will not be available or may not work as expected.  

However, any environment variables referenced in the form of `$var` or `${var}` will expanded. In addition, bash-style
multiline commands (i.e. non-quoted backslashes at the end of a line that indicate a line continuation) will also be 
supported. 

Do not set the Cloud Run region you'd like to deploy to through the `--region` flag in the `gcloud run` commands.
Instead, as mentioned above, do so by setting the `run/region` gcloud property.

### README location
If you wish to parse a README file located somewhere other than the root directory, you can include the README's location
in a `config.yaml` file in the root directory, using the key `readme`. You can specify an absolute directory, or you can simply
specify a directory relative to the sample's directory.

For example, if the README is in the parent directory of the sample:
```text
readme: ../README.md
```
