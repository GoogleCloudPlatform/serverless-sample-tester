package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

const cloudRunServiceNameRandSuffixLen = 10

// gcloudCommandBuildcreates an exec.Cmd that calls the external gcloud executable. The exec.Cmd's name is set to
// "gcloud" and the args are the ones provided in addition to a project flag and quiet flag.
func gcloudCommandBuild(arg []string) *exec.Cmd {
	var gcpProject string
	if sample == nil {
		gcpProject = os.Getenv("GOOGLE_CLOUD_PROJECT")
	} else {
		gcpProject = sample.googleCloudProject
	}

	arg = append(arg, fmt.Sprintf("--project=%s", gcpProject), "--quiet")
	cmd := exec.Command("gcloud", arg...)

	return cmd
}

// execCommand executes an exec.Cmd. It redirects the commands stderr to this program's stderr and returns the output
// in the form of a string.
func execCommand(cmd *exec.Cmd) string {
	if sample == nil {
		cmd.Dir = sampleDir
	} else {
		cmd.Dir = sample.dir
	}

	cmd.Stderr = os.Stderr

	log.Println("Executing ", cmd.String())

	b, err := cmd.Output()
	if err != nil {
		fmt.Println(string(b))
		log.Panicf("Error with exec cmd: %v\n", err)
	}

	result := string(b)
	return strings.TrimSpace(result)
}
