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
	"os/exec"
	"regexp"
	"strings"
)

func parseCloudBuildConfig(filename, serviceName, gcrURL string, substitutions map[string]string) (Lifecycle, error) {
	config := make(map[string]interface{})

	buildConfigBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("[lifecycle.parseCloudBuildConfig] reading Cloud Build config file: %w", err)
	}

	err = yaml.Unmarshal(buildConfigBytes, &config)
	if err != nil {
		return nil, fmt.Errorf("[lifecycle.parseCloudBuildConfig] unmarshaling Cloud Build config file: %w", err)
	}


	// Replace Cloud Run service names and Cloud Container Registry URLs
	for stepIndex := range config["steps"].([]interface{}) {
		runCommand := false
		lastArgIndex := -1

		for argIndex := range config["steps"].([]interface{})[stepIndex].(map[interface{}]interface{})["args"].([]interface{}) {
			arg := config["steps"].([]interface{})[stepIndex].(map[interface{}]interface{})["args"].([]interface{})[argIndex].(string)

			if strings.Contains(arg, "run") {
				runCommand = true
			}

			if !strings.Contains(arg, "--") {
				lastArgIndex = argIndex
			}

			arg = replaceGCRURL(arg, gcrURL)
			config["steps"].([]interface{})[stepIndex].(map[interface{}]interface{})["args"].([]interface{})[argIndex] = arg
		}

		if runCommand && lastArgIndex != -1 {
			config["steps"].([]interface{})[stepIndex].(map[interface{}]interface{})["args"].([]interface{})[lastArgIndex] = serviceName
		}
	}

	configMarshalBytes, err := yaml.Marshal(&config)
	if err != nil {
		return nil, fmt.Errorf("[lifecycle.parseCloudBuildConfig] marshaling modified Cloud Build config: %w", err)
	}

	tempBuildConfigFile, err := util.CreateTempFile()
	if err != nil {
		return nil, fmt.Errorf("[lifecycle.parseCloudBuildConfig] creating temporary file: %w", err)
	}

	if _, err := tempBuildConfigFile.Write(configMarshalBytes); err != nil {
		return nil, fmt.Errorf("[lifecycle.parseCloudBuildConfig] writing to temporary file: %w", err)
	}
	if err := tempBuildConfigFile.Close(); err != nil {
		return nil, fmt.Errorf("[lifecycle.parseCloudBuildConfig] closing temporary file: %w", err)
	}

	return buildCloudBuildConfigLifecycle(tempBuildConfigFile.Name(), substitutions), nil
}

func buildCloudBuildConfigLifecycle(buildConfigFilename string, substitutions map[string]string) Lifecycle {
	a := append(util.GcloudCommonFlags, "builds", "submit",
		fmt.Sprintf("--config=%s", buildConfigFilename))

	subsitutions, empty := substitutionsString(substitutions)
	if !empty {
		a = append(a, fmt.Sprintf("--substitutions=%s", subsitutions))
	}

	return Lifecycle{exec.Command("gcloud", a...)}
}

// replaceServiceName takes a terminal command string as input and replaces the URL of a container image stored in the
// GCP Container Registry with the given URL.
func replaceGCRURL(commandStr string, gcrURL string) string {
	re := regexp.MustCompile(`gcr.io/.+/\S+`)
	return re.ReplaceAllString(commandStr, gcrURL)
}

func substitutionsString(m map[string]string) (string, bool) {
	if len(m) == 0 {
		return "", true
	}

	var subs []string
	for k, v := range m {
		subs = append(subs, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(subs, ","), false
}
