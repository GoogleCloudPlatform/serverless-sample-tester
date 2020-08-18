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

package lifecycle

import (
	"fmt"
	"github.com/GoogleCloudPlatform/serverless-sample-tester/internal/util"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

// runRegionSubstitution is the substitution used to specify which Cloud Run region Cloud Build configs will deploy
// samples to.
const runRegionSubstitution = "_SST_RUN_REGION"

// getCloudBuildConfigLifecycle returns a Lifecycle for the executing the provided Cloud Build config file. It creates
// and uses a temporary copy of the file where it replaces the Cloud Run service names and Container Registry tags with
// the provided inputs. It provides also passes in the provided substitutions as well a runRegionSubstitution with the
// provided region. Also returns a function that removes the temp file created while making Lifecycle. This function
// should be called after Lifecycle is done executing.
func getCloudBuildConfigLifecycle(filename, serviceName, gcrURL, runRegion string, substitutions map[string]string) (Lifecycle, func(), error) {
	config := make(map[string]interface{})

	buildConfigBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("[lifecycle.parseCloudBuildConfig] reading Cloud Build config file: %w", err)
	}

	err = yaml.Unmarshal(buildConfigBytes, &config)
	if err != nil {
		return nil, nil, fmt.Errorf("[lifecycle.parseCloudBuildConfig] unmarshaling Cloud Build config file: %w", err)
	}

	// Replace Cloud Run service names and Container Registry URLs
	for stepIndex := range config["steps"].([]interface{}) {
		var args []string
		for argIndex := range config["steps"].([]interface{})[stepIndex].(map[interface{}]interface{})["args"].([]interface{}) {
			arg := config["steps"].([]interface{})[stepIndex].(map[interface{}]interface{})["args"].([]interface{})[argIndex].(string)
			arg = gcrURLRegexp.ReplaceAllString(arg, gcrURL)

			args = append(args, arg)
		}

		prog := config["steps"].([]interface{})[stepIndex].(map[interface{}]interface{})["name"].(string)
		err := replaceServiceName(prog, args, serviceName)
		if err != nil {
			return nil, nil, fmt.Errorf("[lifecycle.parseCloudBuildConfig] replacing Cloud Run service name in Cloud Build config step args: %w", err)
		}

		config["steps"].([]interface{})[stepIndex].(map[interface{}]interface{})["args"] = args
	}

	configMarshalBytes, err := yaml.Marshal(&config)
	if err != nil {
		return nil, nil, fmt.Errorf("[lifecycle.parseCloudBuildConfig] marshaling modified Cloud Build config: %w", err)
	}

	tempBuildConfigFile, err := ioutil.TempFile("", "example")
	if err != nil {
		return nil, nil, fmt.Errorf("[lifecycle.parseCloudBuildConfig] creating Temp File: %w\n", err)
	}
	cleanup := func() {
		err := os.Remove(tempBuildConfigFile.Name())
		if err != nil {
			log.Printf("Error removing Temp File for Cloud Build config: %v\n", err)
		}
	}

	if _, err := tempBuildConfigFile.Write(configMarshalBytes); err != nil {
		return nil, cleanup, fmt.Errorf("[lifecycle.parseCloudBuildConfig] writing to temporary file: %w", err)
	}
	if err := tempBuildConfigFile.Close(); err != nil {
		return nil, cleanup, fmt.Errorf("[lifecycle.parseCloudBuildConfig] closing temporary file: %w", err)
	}

	return buildCloudBuildConfigLifecycle(tempBuildConfigFile.Name(), runRegion, substitutions), cleanup, nil
}

// buildCloudBuildConfigLifecycle returns a Lifecycle with a single command that calls gcloud builds subit and passes
// in the provided Cloud Build config file. It also adds a `--substitutions` flag according to the substitutions
// provided and adds a substitution for the Cloud Run region with the name runRegionSubstitution and value provided.
func buildCloudBuildConfigLifecycle(buildConfigFilename, runRegion string, substitutions map[string]string) Lifecycle {
	a := append(util.GcloudCommonFlags, "builds", "submit",
		fmt.Sprintf("--config=%s", buildConfigFilename))

	subsitutions := substitutionsString(substitutions, runRegion)
	a = append(a, fmt.Sprintf("--substitutions=%s", subsitutions))

	return Lifecycle{exec.Command("gcloud", a...)}
}

// substitutionsString takes a string to string map and converts it into an argument for the `gcloud builds submit`
// `--config` file. It treats the keys in the map as the substitutions and the values as the substitution values. It
// also adds a substitution for the Cloud Run region with the name runRegionSubstitution and value provided.
func substitutionsString(m map[string]string, runRegion string) string {
	var subs []string
	subs = append(subs, fmt.Sprintf("%s=%s", runRegionSubstitution, runRegion))

	for k, v := range m {
		subs = append(subs, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(subs, ",")
}
