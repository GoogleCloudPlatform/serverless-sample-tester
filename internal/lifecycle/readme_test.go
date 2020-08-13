package lifecycle

import (
	"bufio"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func setEnv(e map[string]string) error {
	for k, v := range e {
		err := os.Setenv(k, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func unsetEnv(e map[string]string) error {
	for k, _ := range e {
		err := os.Unsetenv(k)
		if err != nil {
			return err
		}
	}

	return nil
}

func equalError(a, b error) bool {
	if a == nil {
		return b == nil
	}
	if b == nil {
		return a == nil
	}
	return a.Error() == b.Error()
}

type test struct {
	in                   string
	codeBlocks           []codeBlock
	cmds                 []*exec.Cmd
	toCommandsErr        error
	extractLifecycleErr  error
	extractCodeBlocksErr error
	env                  map[string]string
	serviceName          string
	gcrURL               string
}

var tests = []test{
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
	{
		in: "```\n" +
			"echo hello world\n" +
			"```\n",
		codeBlocks:          nil,
		cmds:                nil,
		extractLifecycleErr: errNoREADMECodeBlocksFound,
	},
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
	//{ this test breaks right now (issue #3)
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
	for i, tt := range tests {
		err := setEnv(tt.env)
		if err != nil {
			t.Errorf("#%d: setEnv: %#v", i, err)

			err = unsetEnv(tt.env)
			if err != nil {
				t.Errorf("#%d: unsetEnv: %#v", i, err)
			}

			continue
		}

		equalE := true
		var cmds []*exec.Cmd
		for j, codeBlock := range tt.codeBlocks {
			h, err := codeBlock.toCommands(tt.serviceName, tt.gcrURL)
			equalE = equalE && equalError(err, tt.toCommandsErr)
			if !equalE {
				t.Errorf("#%d.%d: error mismatch: %v, want %v", i, j, err, tt.toCommandsErr)
			}

			cmds = append(cmds, h...)
		}

		if equalE && !reflect.DeepEqual(cmds, tt.cmds) {
			t.Errorf("#%d: result mismatch\nhave: %#+v\nwant: %#+v", i, cmds, tt.cmds)
		}

		err = unsetEnv(tt.env)
		if err != nil {
			t.Errorf("#%d: unsetEnv: %#v", i, err)
		}
	}
}

func TestExtractLifecycle(t *testing.T) {
	for i, tt := range tests {
		err := setEnv(tt.env)
		if err != nil {
			t.Errorf("#%d: setEnv: %#v", i, err)

			err = unsetEnv(tt.env)
			if err != nil {
				t.Errorf("#%d: unsetEnv: %#v", i, err)
			}

			continue
		}

		s := bufio.NewScanner(strings.NewReader(tt.in))
		var c []*exec.Cmd
		c, err = extractLifecycle(s, tt.serviceName, tt.gcrURL)

		eE := equalError(err, tt.extractLifecycleErr)
		if !eE {
			t.Errorf("#%d: error mismatch: %v, want %v", i, err, tt.extractLifecycleErr)
		}

		if eE && !reflect.DeepEqual(c, tt.cmds) {
			t.Errorf("#%d: result mismatch\nhave: %#+v\nwant: %#+v", i, c, tt.cmds)
		}

		err = unsetEnv(tt.env)
		if err != nil {
			t.Errorf("#%d: unsetEnv: %#v", i, err)
		}
	}
}

func TestExtractCodeBlocks(t *testing.T) {
	for i, tt := range tests {
		s := bufio.NewScanner(strings.NewReader(tt.in))

		h, err := extractCodeBlocks(s)
		if !equalError(err, tt.extractCodeBlocksErr) {
			t.Errorf("#%d: error mismatch: %v, want %v", i, err, tt.extractCodeBlocksErr)
			continue
		}

		if !reflect.DeepEqual(h, tt.codeBlocks) {
			t.Errorf("#%d: result mismatch\nhave: %#+v\nwant: %#+v", i, h, tt.codeBlocks)
			continue
		}
	}
}
