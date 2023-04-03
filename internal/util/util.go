package util

import (
	"fmt"
	"strings"
)

func ConvertRequestToResourceName(requestName string) string {
	splitName := strings.SplitN(requestName, "/", 2)
	resourceType, name := splitName[0], splitName[1]
	resourceName := fmt.Sprintf("resources/%s/types/%s", name, resourceType)

	return resourceName
}

func ConvertServiceToResourceName(serviceName string) string {
	resourceName := fmt.Sprintf("resources/%s/types/%s", serviceName, RESOURCE_TYPE_SERVICE)

	return resourceName
}

func ConvertWorkflfowToWorkflowResourceName(resourceName string) string {
	resourceWorkflowId := fmt.Sprintf("%s/workflow", resourceName)

	return resourceWorkflowId
}
