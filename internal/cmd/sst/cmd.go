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

package sst

import (
	"fmt"
	"github.com/GoogleCloudPlatform/serverless-sample-tester/internal/sample"
	"github.com/GoogleCloudPlatform/serverless-sample-tester/internal/util"
	"github.com/spf13/cobra"
	"log"
	"path/filepath"
)

// Root is responsible for the root command. It handles the application flow.
func Root(cmd *cobra.Command, args []string) error {
	// Parse sample directory from command line argument
	sampleDir, err := filepath.Abs(filepath.Dir(args[0]))
	if err != nil {
		return err
	}
	util.SetCommandsDir(sampleDir)

	log.Println("Setting up configuration values")
	s, err := sample.NewSample(sampleDir)
	if err != nil {
		return err
	}

	log.Println("Loading test endpoints")
	swagger := util.LoadTestEndpoints()

	log.Println("Building and deploying sample to Cloud Run")
	err = s.BuildDeployLifecycle.Execute()
	defer s.Service.Delete()
	defer s.DeleteCloudContainerImage()
	if err != nil {
		return err
	}

	log.Println("Getting identity token for gcloud auhtorized account")
	var identToken string
	identToken, err = util.ExecCommand(util.GcloudCommandBuild(
		"auth",
		"print-identity-token",
	))
	if err != nil {
		return err
	}

	log.Println("Checking endpoints for expected results")
	serviceURL, err := s.Service.URL()
	if err != nil {
		return err
	}

	allTestsPassed, err := util.ValidateEndpoints(serviceURL, &swagger.Paths, identToken)
	if err != nil {
		return err
	}

	if !allTestsPassed {
		return fmt.Errorf("all tests did not pass")
	}

	return nil
}
