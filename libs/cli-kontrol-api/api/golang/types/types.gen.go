// Package types provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version (devel) DO NOT EDIT.
package types

import (
	compose "github.com/compose-spec/compose-go/types"
)

// DevFlowSpec defines model for DevFlowSpec.
type DevFlowSpec struct {
	DockerCompose *[]compose.ServiceConfig `json:"docker-compose,omitempty"`
	ImageLocator  *string                  `json:"image-locator,omitempty"`
	ServiceName   *string                  `json:"service-name,omitempty"`
}

// PostDevFlowJSONRequestBody defines body for PostDevFlow for application/json ContentType.
type PostDevFlowJSONRequestBody = DevFlowSpec
