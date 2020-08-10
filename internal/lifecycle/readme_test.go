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

type toCommandsTest struct {
	in                    codeBlock
	out                   []*exec.Cmd
	err                   error
	env                   map[string]string
	serviceName           string
	gcrURL                string
}

var toCommandsTests = []toCommandsTest{
	{
		in: []string {
			"echo hello world",
		},
		out: []*exec.Cmd{
			exec.Command("echo", "hello", "world"),
		},
	},
	{
		in: []string {
			"echo hello world one",
			"echo hello world two",
		},
		out: []*exec.Cmd{
			exec.Command("echo", "hello", "world", "one"),
			exec.Command("echo", "hello", "world", "two"),
		},
	},
	{
		in: []string {
			"echo multi \\",
			"line command",
		},
		out: []*exec.Cmd{
			exec.Command("echo", "multi", "line", "command"),
		},
	},
	{
		in: []string {
			"echo ${TEST_ENV}",
		},
		out: []*exec.Cmd{
			exec.Command("echo", "hello", "world"),
		},
		env: map[string]string{
			"TEST_ENV": "hello world",
		},
	},
	{
		in: []string {
			"gcloud run services deploy hello_world",
		},
		out: []*exec.Cmd{
			exec.Command("gcloud", "--quiet", "run", "services", "deploy", "unique_service_name"),
		},
		serviceName: "unique_service_name",
	},
	{
		in: []string {
			"gcloud builds submit --tag=gcr.io/hello/world",
		},
		out: []*exec.Cmd{
			exec.Command("gcloud", "--quiet", "builds", "submit", "--tag=gcr.io/unique/tag"),
		},
		gcrURL: "gcr.io/unique/tag",
	},
	{
		in: []string {
			"gcloud run services deploy hello_world --image=gcr.io/hello/world",
		},
		out: []*exec.Cmd{
			exec.Command("gcloud", "--quiet", "run", "services", "deploy", "unique_service_name", "--image=gcr.io/unique/tag"),
		},
		serviceName: "unique_service_name",
		gcrURL: "gcr.io/unique/tag",
	},
	//{ this test breaks right now (issue #3)
	//	in: []string {
	//		"gcloud run services deploy hello_world --image gcr.io/hello/world",
	//	},
	//	out: []*exec.Cmd{
	//		exec.Command("gcloud", "--quiet", "run", "services", "deploy", "unique_service_name", "--image gcr.io/unique/tag"),
	//	},
	//	serviceName: "unique_service_name",
	//	gcrURL: "gcr.io/unique/tag",
	//},
	{
		in: []string {
			"gcloud run services deploy hello_world --image=gcr.io/hello/world --add-cloudsql-instances=${TEST_CLOUD_SQL_CONNECTION}",
		},
		out: []*exec.Cmd{
			exec.Command("gcloud", "--quiet", "run", "services", "deploy", "unique_service_name", "--image=gcr.io/unique/tag", "--add-cloudsql-instances=project:region:instance"),
		},
		env: map[string]string{
			"TEST_CLOUD_SQL_CONNECTION": "project:region:instance",
		},
		serviceName: "unique_service_name",
		gcrURL: "gcr.io/unique/tag",
	},
}

func TestToCommands(t *testing.T) {
	for i, tt := range toCommandsTests {
		err := setEnv(tt.env)
		if err != nil {
			t.Errorf("#%d: setEnv: %#v", i, err)

			err = unsetEnv(tt.env)
			if err != nil {
				t.Errorf("#%d: unsetEnv: %#v", i, err)
			}

			continue
		}

		h, err := tt.in.toCommands(tt.serviceName, tt.gcrURL)
		eE := equalError(err, tt.err)
		if !eE {
			t.Errorf("#%d: error mismatch: %v, want %v", i, err, tt.err)
		}

		if eE && !reflect.DeepEqual(h, tt.out) {
			t.Errorf("#%d: result mismatch\nhave: %#+v\nwant: %#+v", i, h, tt.out)
		}

		err = unsetEnv(tt.env)
		if err != nil {
			t.Errorf("#%d: unsetEnv: %#v", i, err)
		}
	}
}

type extractTest struct {
	in                    string
	out                   []codeBlock
	err                   error
}

var extractTests = []extractTest{
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"echo hello world\n" +
			"```\n",
        out: []codeBlock {
			[]string {
				"echo hello world",
			},
		},
	},
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"echo line one\n" +
			"echo line two\n" +
			"```\n",
		out: []codeBlock {
			[]string {
				"echo line one",
				"echo line two",
			},
		},
	},
	{
		in: "[//]: # ({sst-run-unix})\n" +
			"```\n" +
			"echo multi \\\n" +
			"line command\n" +
			"```\n",
		out: []codeBlock {
			[]string {
				"echo multi \\",
				"line command",
			},
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
		out: []codeBlock {
			[]string {
				"echo build command",
			},
			[]string {
				"echo deploy command",
			},
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
		out: []codeBlock {
			[]string {
				"echo build and deploy command",
			},
		},
	},
	{
		in: "```\n" +
			"echo hello world\n" +
			"```\n",
		out: nil,
	},
}

func TestExtractCodeBlocks(t *testing.T) {
	for i, tt := range extractTests {
		s := bufio.NewScanner(strings.NewReader(tt.in))

		h, err := extractCodeBlocks(s)
		if !equalError(err, tt.err) {
			t.Errorf("#%d: error mismatch: %v, want %v", i, err, tt.err)
			continue
		}

		if !reflect.DeepEqual(h, tt.out) {
			t.Errorf("#%d: result mismatch\nhave: %#+v\nwant: %#+v", i, h, tt.out)
			continue
		}
	}
}
