// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

var (
	sample *Sample

	sampleDir string
	keepContainerImage bool

	allTestsPassed bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "sst [sample-dir]",
		Short: "An end-to-end tester for GCP samples",
		Args:  cobra.ExactArgs(1),
		Run:   root,
	}

	if err := rootCmd.Execute(); err != nil {
		log.Panicf("Error with cobra rootCmd Execution: %v\n", err)
	}

	if allTestsPassed {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

func root(cmd *cobra.Command, args []string) {
	// Parse sample directory from command line argument
	var err error
	sampleDir, err = filepath.Abs(filepath.Dir(args[0]))
	if err != nil {
		log.Fatalf("Error parsing sample direcotry: %v\n", err)
	}

	log.Println("Setting up configuration values")
	sample = newSample(sampleDir)

	log.Println("Loading test endpoints")
	swagger := loadTestEndpoints()

	log.Println("Activating service account")
	execCommand(gcloudCommandBuild([]string{
		"auth",
		"activate-service-account",
		os.ExpandEnv("--key-file=${GOOGLE_APPLICATION_CREDENTIALS}"),
	}))

	log.Println("Building and deploying sample to Cloud Run")
	sample.buildDeployLifecycle.execute()
	defer sample.cloudRunService.delete()
	defer sample.cloudContainerImage.delete()

	log.Println("Getting identity token for service account")
	identToken := execCommand(gcloudCommandBuild([]string{
		"auth",
		"print-identity-token",
	}))

	log.Println("Checking endpoints for expected results")
	allTestsPassed = validateEndpoints(&swagger.Paths, identToken)
}
