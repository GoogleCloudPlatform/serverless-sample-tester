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
	"github.com/getkin/kin-openapi/openapi3"
	"log"
)

const passResponseDescription = "PASS"

// LoadTestEndpoints loads a default test endpoint request (a GET / request expecting a 200 status code) into an
// openapi3.Swagger object (see github.com/getkin/kin-openapi).
func LoadTestEndpoints() *openapi3.Swagger {
	prd := passResponseDescription

	log.Println("Using default test endpoint (GET /)")
	return &openapi3.Swagger{
		Paths: openapi3.Paths{
			"/": &openapi3.PathItem{
				Get: &openapi3.Operation{
					Responses: openapi3.Responses{
						"200": &openapi3.ResponseRef{
							Value: &openapi3.Response{
								ExtensionProps: openapi3.ExtensionProps{},
								Description:    &prd,
							},
						},
					},
				},
			},
		},
	}
}
