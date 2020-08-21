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

package cmd

import (
	"fmt"
	"github.com/GoogleCloudPlatform/serverless-sample-tester/internal/sample"
	"github.com/GoogleCloudPlatform/serverless-sample-tester/internal/util"
	"github.com/spf13/cobra"
	"log"
	"os/exec"
	"path/filepath"
)

var (
	rootCmd = &cobra.Command{
		Use:   "sst [sample-dir]",
		Short: "An end-to-end tester for GCP samples",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse sample directory from command line argument
			sampleDir, err := filepath.Abs(filepath.Dir(args[0]))
			if err != nil {
				return err
			}

			log.Println("Setting up configuration values")
			s, err := sample.NewSample(sampleDir)
			if err != nil {
				return err
			}

			log.Println("Loading test endpoints")
			swagger := util.LoadTestEndpoints()

			log.Println("Building and deploying sample to Cloud Run")
			err = s.BuildDeployLifecycle.Execute(s.Dir)
			defer s.Service.Delete(s.Dir)
			defer s.DeleteCloudContainerImage()
			if err != nil {
				return fmt.Errorf("[cmd.Root] building and deploying sample to Cloud Run: %w", err)
			}

			log.Println("Getting identity token for gcloud auhtorized account")
			var identToken string
			a := append(util.GcloudCommonFlags, "auth", "print-identity-token")
			identToken, err = util.ExecCommand(exec.Command("gcloud", a...), s.Dir)
			if err != nil {
				return fmt.Errorf("[cmd.Root] getting identity token for gcloud auhtorized account: %w", err)
			}

			log.Println("Checking endpoints for expected results")
			serviceURL, err := s.Service.URL(s.Dir)
			if err != nil {
				return fmt.Errorf("[cmd.Root] getting Cloud Run service URL: %w", err)
			}

			log.Println("Validating Cloud Run service endpoints for expected status codes")
			allTestsPassed, err := util.ValidateEndpoints(serviceURL, &swagger.Paths, identToken)
			if err != nil {
				return fmt.Errorf("[cmd.Root] validating Cloud Run service endpoints for expected status codes: %w", err)
			}

			if !allTestsPassed {
				return fmt.Errorf("all tests did not pass")
			}
			return nil
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

// init initializes the tool.
func init() {
	// Initialization goes here
}
