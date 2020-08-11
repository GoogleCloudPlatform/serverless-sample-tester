package util

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

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
