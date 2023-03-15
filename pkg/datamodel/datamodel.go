package datamodel

import (
	controllerPB "github.com/instill-ai/protogen-go/vdp/controller/v1alpha"
)

// Resource defines a resource in the database
type Resource struct {
	Name     string
	State    controllerPB.Resource_State
	Progress int32
}
