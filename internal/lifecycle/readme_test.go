package lifecycle

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

// setEnv takes a map of environment variables to their values and sets the program's environment accordingly.
func setEnv(e map[string]string) error {
	for k, v := range e {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}
	return nil
}

// unsetEnv takes a map of environment variables to their values and unsets the environment variables in the program's
// environment.
func unsetEnv(e map[string]string) error {
	for k, _ := range e {
		if err := os.Unsetenv(k); err != nil {
			return err
		}
	}
	return nil
}

type test struct {
	in                   string            // inupt markdown string
	codeBlocks           []codeBlock       // expected result of extractCodeBlocks on in
	cmds                 []*exec.Cmd       // expected result of toCommands on all codeBlocks and extractLifecycle on in
	toCommandsErr        error             // expected toCommands return error
	extractLifecycleErr  error             // expected extractLifecycle return error
	extractCodeBlocksErr error             // expected extractCodeBlocks return error
	env                  map[string]string // map of environment variables to values for this test
	serviceName          string            // Cloud Run service name that should replace existing names
	gcrURL               string            // Container Registry URL that should replace existing URLs
}

var tests = []test{
	// single code block, single one-line command
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"echo hello world\n" +
			"```\n",
		codeBlocks: []codeBlock{
			[]string{
				"echo hello world",
			},
		},
		cmds: []*exec.Cmd{
			exec.Command("echo", "hello", "world"),
		},
	},

	// code block not closed
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"echo hello world\n",
		codeBlocks:           nil,
		cmds:                 nil,
		extractLifecycleErr:  errCodeBlockNotClosed,
		extractCodeBlocksErr: errCodeBlockNotClosed,
	},

	// code block doesn't start immediately after code tag
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"not start of code block\n" +
			"```\n" +
			"echo hello world\n" +
			"```\n",
		codeBlocks:           nil,
		cmds:                 nil,
		extractLifecycleErr:  errCodeBlockStartNotFound,
		extractCodeBlocksErr: errCodeBlockStartNotFound,
	},

	// EOF immediately after code tag
	{
		in: "instuctions\n" +
			"[//]: # ({sst-run-unix})\n",
		codeBlocks:           nil,
		cmds:                 nil,
		extractLifecycleErr:  errEOFAfterCodeTag,
		extractCodeBlocksErr: errEOFAfterCodeTag,
	},

	// single code block, two one-line commands
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"echo line one\n" +
			"echo line two\n" +
			"```\n",
		codeBlocks: []codeBlock{
			[]string{
				"echo line one",
				"echo line two",
			},
		},
		cmds: []*exec.Cmd{
			exec.Command("echo", "line", "one"),
			exec.Command("echo", "line", "two"),
		},
	},

	// single code block, single multiline command
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"echo multi \\\n" +
			"line command\n" +
			"```\n",
		codeBlocks: []codeBlock{
			[]string{
				"echo multi \\",
				"line command",
			},
		},
		cmds: []*exec.Cmd{
			exec.Command("echo", "multi", "line", "command"),
		},
	},

	// line cont char but code block closes at next line
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"echo multi \\\n" +
			"```\n",
		codeBlocks: []codeBlock{
			[]string{
				"echo multi \\",
			},
		},
		cmds:                nil,
		toCommandsErr:       errCodeBlockEndAfterLineCont,
		extractLifecycleErr: errCodeBlockEndAfterLineCont,
	},

	// two code blocks, one single-line command in each, with markdown instructions in the middle
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"echo build command\n" +
			"```\n" +
			"markdown instructions\n" +
			"[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"echo deploy command\n" +
			"```\n",
		codeBlocks: []codeBlock{
			[]string{
				"echo build command",
			},
			[]string{
				"echo deploy command",
			},
		},
		cmds: []*exec.Cmd{
			exec.Command("echo", "build", "command"),
			exec.Command("echo", "deploy", "command"),
		},
	},

	// two code blocks, but only one is annotated with code tag
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"echo build and deploy command\n" +
			"```\n" +
			"markdown instructions\n" +
			"```\n" +
			"echo irrelevant command\n" +
			"```\n",
		codeBlocks: []codeBlock{
			[]string{
				"echo build and deploy command",
			},
		},
		cmds: []*exec.Cmd{
			exec.Command("echo", "build", "and", "deploy", "command"),
		},
	},

	// one code block, but not annotated with code tag
	{
		in: "```\n" +
			"echo hello world\n" +
			"```\n",
		codeBlocks:          nil,
		cmds:                nil,
		extractLifecycleErr: errNoREADMECodeBlocksFound,
	},

	// expand environment variable test
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"echo ${TEST_ENV}\n" +
			"```\n",
		codeBlocks: []codeBlock{
			[]string{
				"echo ${TEST_ENV}",
			},
		},
		cmds: []*exec.Cmd{
			exec.Command("echo", "hello", "world"),
		},
		env: map[string]string{
			"TEST_ENV": "hello world",
		},
	},

	// replace Cloud Run service name with provided name
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"gcloud run services deploy hello_world\n" +
			"```\n",
		codeBlocks: []codeBlock{
			[]string{
				"gcloud run services deploy hello_world",
			},
		},
		cmds: []*exec.Cmd{
			exec.Command("gcloud", "--quiet", "run", "services", "deploy", "unique_service_name"),
		},
		serviceName: "unique_service_name",
	},

	// replace Container Registry URL with provided URL
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"gcloud builds submit --tag=gcr.io/hello/world\n" +
			"```\n",
		codeBlocks: []codeBlock{
			[]string{
				"gcloud builds submit --tag=gcr.io/hello/world",
			},
		},
		cmds: []*exec.Cmd{
			exec.Command("gcloud", "--quiet", "builds", "submit", "--tag=gcr.io/unique/tag"),
		},
		gcrURL: "gcr.io/unique/tag",
	},

	// replace multiline GCR URL with provided URL
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"gcloud builds submit --tag=gcr.io/hello/\\\n" +
			"world\n" +
			"```\n",
		codeBlocks: []codeBlock{
			[]string{
				"gcloud builds submit --tag=gcr.io/hello/\\",
				"world",
			},
		},
		cmds: []*exec.Cmd{
			exec.Command("gcloud", "--quiet", "builds", "submit", "--tag=gcr.io/unique/tag"),
		},
		gcrURL: "gcr.io/unique/tag",
	},

	// replace Cloud Run service name and GCR URL with provided inputs
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"gcloud run services deploy hello_world --image=gcr.io/hello/world\n" +
			"```\n",
		codeBlocks: []codeBlock{
			[]string{
				"gcloud run services deploy hello_world --image=gcr.io/hello/world",
			},
		},
		cmds: []*exec.Cmd{
			exec.Command("gcloud", "--quiet", "run", "services", "deploy", "unique_service_name", "--image=gcr.io/unique/tag"),
		},
		serviceName: "unique_service_name",
		gcrURL:      "gcr.io/unique/tag",
	},

	// replace Cloud Run service name and GCR URL with `--image url` syntax
	// this test breaks right now (issue #3)
	//{
	//	in: "[//]: # ({sst-run-unix})\n" +
	//		"```\n" +
	//		"gcloud run services deploy hello_world --image gcr.io/hello/world\n" +
	//		"```\n",
	//	codeBlocks: []codeBlock {
	//		[]string {
	//			"gcloud run services deploy hello_world --image gcr.io/hello/world",
	//		},
	//	},
	//	cmds: []*exec.Cmd{
	//		exec.Command("gcloud", "--quiet", "run", "services", "deploy", "unique_service_name", "--image", "gcr.io/unique/tag"),
	//	},
	//	serviceName: "unique_service_name",
	//	gcrURL: "gcr.io/unique/tag",
	//},

	// replace Cloud Run service name and GCR URL with provided inputs and expand environment variables
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"gcloud run services deploy hello_world --image=gcr.io/hello/world --add-cloudsql-instances=${TEST_CLOUD_SQL_CONNECTION}\n" +
			"```\n",
		codeBlocks: []codeBlock{
			[]string{
				"gcloud run services deploy hello_world --image=gcr.io/hello/world --add-cloudsql-instances=${TEST_CLOUD_SQL_CONNECTION}",
			},
		},
		cmds: []*exec.Cmd{
			exec.Command("gcloud", "--quiet", "run", "services", "deploy", "unique_service_name", "--image=gcr.io/unique/tag", "--add-cloudsql-instances=project:region:instance"),
		},
		env: map[string]string{
			"TEST_CLOUD_SQL_CONNECTION": "project:region:instance",
		},
		serviceName: "unique_service_name",
		gcrURL:      "gcr.io/unique/tag",
	},

	// replace Cloud Run service name provided name in command with multiline arguments
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"gcloud run services update hello_world --add-cloudsql-instances=\\\n" +
			"project:region:instance\n" +
			"```\n",
		codeBlocks: []codeBlock{
			[]string{
				"gcloud run services update hello_world --add-cloudsql-instances=\\",
				"project:region:instance",
			},
		},
		cmds: []*exec.Cmd{
			exec.Command("gcloud", "--quiet", "run", "services", "update", "unique_service_name", "--add-cloudsql-instances=project:region:instance"),
		},
		serviceName: "unique_service_name",
		gcrURL:      "gcr.io/unique/tag",
	},

	// replace Cloud Run service name provided name and expand environment variables in command with multiline arguments
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"gcloud run services update hello_world --add-cloudsql-instances=\\\n" +
			"${TEST_CLOUD_SQL_CONNECTION}\n" +
			"```\n",
		codeBlocks: []codeBlock{
			[]string{
				"gcloud run services update hello_world --add-cloudsql-instances=\\",
				"${TEST_CLOUD_SQL_CONNECTION}",
			},
		},
		cmds: []*exec.Cmd{
			exec.Command("gcloud", "--quiet", "run", "services", "update", "unique_service_name", "--add-cloudsql-instances=project:region:instance"),
		},
		env: map[string]string{
			"TEST_CLOUD_SQL_CONNECTION": "project:region:instance",
		},
		serviceName: "unique_service_name",
		gcrURL:      "gcr.io/unique/tag",
	},
}

func TestToCommands(t *testing.T) {
	for i, tc := range tests {
		if err := setEnv(tc.env); err != nil {
			t.Errorf("#%d: setEnv: %v", i, err)

			if err = unsetEnv(tc.env); err != nil {
				t.Errorf("#%d: unsetEnv: %v", i, err)
			}

			continue
		}

		matchE := true
		var cmds []*exec.Cmd
		for j, codeBlock := range tc.codeBlocks {
			h, err := codeBlock.toCommands(tc.serviceName, tc.gcrURL)
			matchE = matchE && errors.Is(err, tc.toCommandsErr)
			if !matchE {
				t.Errorf("#%d.%d: error mismatch\nwant: %v\ngot: %v", i, j, tc.toCommandsErr, err)
			}

			cmds = append(cmds, h...)
		}

		if matchE && !reflect.DeepEqual(cmds, tc.cmds) {
			t.Errorf("#%d: result mismatch\nwant: %#+v\ngot: %#+v", i, tc.cmds, cmds)
		}

		if err := unsetEnv(tc.env); err != nil {
			t.Errorf("#%d: unsetEnv: %v", i, err)
		}
	}
}

func TestExtractLifecycle(t *testing.T) {
	for i, tc := range tests {
		if err := setEnv(tc.env); err != nil {
			t.Errorf("#%d: setEnv: %v", i, err)

			if err = unsetEnv(tc.env); err != nil {
				t.Errorf("#%d: unsetEnv: %v", i, err)
			}

			continue
		}

		s := bufio.NewScanner(strings.NewReader(tc.in))
		var cmds []*exec.Cmd
		cmds, err := extractLifecycle(s, tc.serviceName, tc.gcrURL)

		mE := errors.Is(err, tc.extractLifecycleErr)
		if !mE {
			t.Errorf("#%d: error mismatch\nwant: %v\ngot: %v", i, tc.extractLifecycleErr, err)
		}

		if mE && !reflect.DeepEqual(cmds, tc.cmds) {
			t.Errorf("#%d: result mismatch\nwant: %#+v\ngot: %#+v", i, tc.cmds, cmds)
		}

		if err = unsetEnv(tc.env); err != nil {
			t.Errorf("#%d: unsetEnv: %v", i, err)
		}
	}
}

func TestExtractCodeBlocks(t *testing.T) {
	for i, tc := range tests {
		s := bufio.NewScanner(strings.NewReader(tc.in))

		codeBlocks, err := extractCodeBlocks(s)
		if !errors.Is(err, tc.extractCodeBlocksErr) {
			t.Errorf("#%d: error mismatch\nwant: %v\ngot: %v", i, tc.extractCodeBlocksErr, err)
			continue
		}

		if !reflect.DeepEqual(codeBlocks, tc.codeBlocks) {
			t.Errorf("#%d: result mismatch\nwant: %#+v\ngot: %#+v", i, tc.codeBlocks, codeBlocks)
			continue
		}
	}
}
