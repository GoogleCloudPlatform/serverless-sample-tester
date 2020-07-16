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
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var identityToken string

// validateEndpoints tests all paths (represented by openapi3.Paths) with all HTTP methods and given response bodies
// and make sure they respond with the expected status code. Returns a success bool based on whether all the tests
// passed.
func validateEndpoints(paths *openapi3.Paths, identTok string) bool {
	identityToken = identTok

	s := true
	for endpoint, pathItem := range *paths {
		log.Printf("Testing %s endpoint\n", endpoint)
		s = s && validateEndpointOperation(pathItem.Connect, endpoint, http.MethodConnect)
		s = s && validateEndpointOperation(pathItem.Delete, endpoint, http.MethodDelete)
		s = s && validateEndpointOperation(pathItem.Get, endpoint, http.MethodGet)
		s = s && validateEndpointOperation(pathItem.Head, endpoint, http.MethodHead)
		s = s && validateEndpointOperation(pathItem.Options, endpoint, http.MethodOptions)
		s = s && validateEndpointOperation(pathItem.Patch, endpoint, http.MethodPatch)
		s = s && validateEndpointOperation(pathItem.Post, endpoint, http.MethodPost)
		s = s && validateEndpointOperation(pathItem.Put, endpoint, http.MethodPut)
		s = s && validateEndpointOperation(pathItem.Trace, endpoint, http.MethodTrace)
	}

	return s
}

// validateEndpointOperation validates a single endpoint and a single HTTP method, and ensures that the request --
// including the provided sample request body -- elicits the expected status code.
func validateEndpointOperation(operation *openapi3.Operation, endpoint string, httpMethod string) bool {
	if operation == nil {
		return true
	}
	log.Printf("%s %s\n", httpMethod, endpoint)

	if operation.RequestBody == nil {
		log.Println("Empty request body")
		reqBodyReader := strings.NewReader("")

		return makeTestRequest(httpMethod, endpoint, "", reqBodyReader, operation)
	}

	reqBodies := operation.RequestBody.Value.Content
	allTestsPassed := true
	for mimeType, mediaType := range reqBodies {
		reqBodyStr := mediaType.Example.(string)
		log.Printf("%s: %s", mimeType, reqBodyStr)

		reqBodyReader := strings.NewReader(reqBodyStr)
		allTestsPassed = allTestsPassed && makeTestRequest(httpMethod, endpoint, mimeType, reqBodyReader, operation)
	}

	return allTestsPassed
}

// makeTestRequest returns a success bool based on whether the returned status code
// was included in the provided openapi3.Operation expected responses.
func makeTestRequest(httpMethod, endpoint, mimeType string, reqBodyReader *strings.Reader, operation *openapi3.Operation) bool {
	client := &http.DefaultClient

	req, err := http.NewRequest(httpMethod, s.service.getURL()+endpoint, reqBodyReader)
	if err != nil {
		log.Panicf("Error creating http request: %v\n", err)
	}

	req.Header.Add("Authorization", "Bearer "+identityToken)
	req.Header.Add("content-type", mimeType)

	resp, err := (*client).Do(req)
	if err != nil {
		log.Panicf("Error executing http request: %v\n", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panicf("Error reading http response body: %v\n", err)
	}
	defer resp.Body.Close()

	statusCode := strconv.Itoa(resp.StatusCode)
	log.Printf("Status code: %s\n", statusCode)

	if val, ok := operation.Responses[statusCode]; ok {
		log.Printf("Response description: %s\n", *val.Value.Description)
		return true
	} else {
		log.Println("Unknown response description: FAIL")
		log.Println("Dumping response body")
		fmt.Println(string(body))
		return false
	}
}
