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

package util

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// tempFiles is a slice of pointers to temporary files created throughout the course of the program. Users should call
// RemoveTempFiles at the end of the program.
var tempFiles []*os.File

func CreateTempFile() (*os.File, error) {
	tempFile, err := ioutil.TempFile("", "example")
	if err != nil {
		return tempFile, fmt.Errorf("[util.CreateTempFile] creating temp file: %w\n", err)
	}

	tempFiles = append(tempFiles, tempFile)
	return tempFile, nil
}

func RemoveTempFiles() error {
	for _, tempFile := range tempFiles {
		err := os.Remove(tempFile.Name())
		if err != nil {
			log.Printf("Error removing Temp File: %v\n", err)
		}
	}

	return nil
}
