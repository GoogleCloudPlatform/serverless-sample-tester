package main

import (
	"github.com/getkin/kin-openapi/openapi3"
	"log"
)

const passResponseDescription = "PASS"

// loadTestEndpoints loads a default test endpoint request (a GET / request expecting a 200 status code) into an
// openapi3.Swagger object (see github.com/getkin/kin-openapi).
func loadTestEndpoints() *openapi3.Swagger {
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
