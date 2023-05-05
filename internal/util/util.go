package util

import (
	"fmt"
	"strings"
)

func ConvertRequestToResourceName(requestName string, uid string) string {
	resourceType := strings.SplitN(requestName, "/", 2)[0]
	resourceName := fmt.Sprintf("resources/%s/types/%s", uid, resourceType)

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
