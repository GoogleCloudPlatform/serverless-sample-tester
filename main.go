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
	s *sample

	sampleDir string

	allTestsPassed bool

	err error
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "sst [sample-dir]",
		Short: "An end-to-end tester for GCP samples",
		Args:  cobra.ExactArgs(1),
		Run:   root,
	}

	if e := rootCmd.Execute(); e != nil {
		log.Fatalf("Error with cobra rootCmd Execution: %v\n", err)
	}

	if !allTestsPassed || err != nil {
		log.Fatalf("Error occured in the exectuion of this program: %v", err)
	}
}

func root(cmd *cobra.Command, args []string) {
	// Parse sample directory from command line argument
	sampleDir, err = filepath.Abs(filepath.Dir(args[0]))
	if err != nil {
		return
	}

	log.Println("Setting up configuration values")
	s, err = newSample(sampleDir)
	if err != nil {
		return
	}

	log.Println("Loading test endpoints")
	swagger := loadTestEndpoints()

	log.Println("Activating service account")
	_, err = execCommand(gcloudCommandBuild([]string{
		"auth",
		"activate-service-account",
		os.ExpandEnv("--key-file=${GOOGLE_APPLICATION_CREDENTIALS}"),
	}))
	if err != nil {
		return
	}

	log.Println("Building and deploying sample to Cloud Run")
	err = s.buildDeployLifecycle.execute()
	defer s.service.delete()
	defer s.container.delete()
	if err != nil {
		return
	}

	log.Println("Getting identity token for service account")
	var identToken string
	identToken, err = execCommand(gcloudCommandBuild([]string{
		"auth",
		"print-identity-token",
	}))
	if err != nil {
		return
	}

	log.Println("Checking endpoints for expected results")
	allTestsPassed, err = validateEndpoints(&swagger.Paths, identToken)
}
