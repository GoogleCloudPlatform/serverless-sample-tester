# Serverless Sample Tester

This is an end-to-end framework that deploys Google Cloud Platform samples to
Cloud Run and ensures that they perform as expected when deployed.

This project’s primary intended users are developers looking to test GCP
samples. In order to streamline the development workflow, this project is
focused on being an end-to-end testing framework that specifically targets
identifying bugs that arise when samples are deployed to Cloud Run. It will:

1. Deploy samples to Cloud Run
1. Check deployed service for expected responses
1. Return logs of health check service’s logs if any tests failed
1. Clean up any created resources as part of previous processes

## Build

```bash
go build -o sst main.go
```

## Usage

```bash
./sst [sample-dir]
```

Make sure to authorize the gcloud SDK and set a default project and Cloud Run region before running this program. A
default Cloud Run region can be set by setting the `run/region` gcloud property.

### README parsing
If you'd like, make sure to include the following comment code tag in your README to customize how the program should
build and deploy your sample:

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

When parsing the README for custom build and deploy commands, the serverless sample tester will include any commands
inside a code fence that is immediately preceded by a line containing `{sst-run-unix}`. You can use Markdown syntax
(e.g. `[//]: # ({sst-run-unix})`) to include this line without making it visible when rendered.

The parsed commands will not be run through a shell, meaning that the program will not perform any expansions,
pipelines, redirections or any other functions that shells are responsible for. This also means that popular shell
builtin commands like `cd`, `export`, and `echo` will not be available or may not work as expected.  

However, any environment variables referenced in the form of `$var` or `${var}` will expanded. In addition, bash-style
multiline commands (i.e. non-quoted backslashes at the end of a line that indicate a line continuation) will also be 
supported. 

Do not set the Cloud Run region you'd like to deploy to through the `--region` flag in the `gcloud run` commands.
Instead, as mentioned above, do so by setting the `run/region` gcloud property.

In addition to setting a default Cloud Run region, make sure to deploy to the fully managed platform on Cloud Run. You
can achieve this by setting the `run/platform` gcloud property to `managed` or passing in the `--platform=managed` flag
to your `gcloud run` commands.

If comment code tags aren't added to your README, the program will fall back to reasonable defaults to build and deploy
your sample to Cloud Run based on whether your sample is java-based and doesn't have a Dockerfile or isn't.
