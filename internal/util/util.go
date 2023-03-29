package util

import (
	"fmt"
	"strings"
)

func ConvertModelToResourceName(modelInstanceName string) string {
	splitName := strings.SplitN(modelInstanceName, "/", 4)
	modelType, modelID, modelInstanceID := splitName[0], splitName[1], splitName[3]
	resourceName := fmt.Sprintf("resources/%s_%s/types/%s", modelID, modelInstanceID, modelType)

	return resourceName
}

func ConvertConnectorToResourceName(connectorName string) string {
	splitName := strings.SplitN(connectorName, "/", 2)
	connectorType, name := splitName[0], splitName[1]
	resourceName := fmt.Sprintf("resources/%s/types/%s", name, connectorType)

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
