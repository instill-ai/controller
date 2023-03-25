package util

import (
	"fmt"
	"strings"
)

func ConvertModelToResourceName(modelInstanceName string) string {
	splitName := strings.SplitN(modelInstanceName, "/", 4)
	modelID, modelInstanceID := splitName[1], splitName[3]
	resourceName := fmt.Sprintf("resources/%s_%s/types/model", modelID, modelInstanceID)

	return resourceName
}

func ConvertConnectorToResourceName(connectorName string) string {
	splitName := strings.SplitN(connectorName, "/", 2)
	connectorType, name := splitName[0], splitName[1]
	resourceName := fmt.Sprintf("resources/%s/types/%s", name, connectorType)

	return resourceName
}

func ConvertServiceToResourceName(serviceName string) string {
	resourceServiceName := fmt.Sprintf("resources/%s/types/service", serviceName)

	return resourceServiceName
}

func ConvertWorkflfowToWorkflowResourceName(resourceName string) string {
	resourceWorkflowID := fmt.Sprintf("%s/workflow", resourceName)

	return resourceWorkflowID
}
