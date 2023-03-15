package handler

import (
	"context"
	"strings"

	"github.com/instill-ai/controller/pkg/service"

	controllerPB "github.com/instill-ai/protogen-go/vdp/controller/v1alpha"
	healthcheckPB "github.com/instill-ai/protogen-go/vdp/healthcheck/v1alpha"
)

type PrivateHandler struct {
	controllerPB.UnimplementedControllerPrivateServiceServer
	service service.Service
}

func NewPrivateHandler(s service.Service) controllerPB.ControllerPrivateServiceServer {
	return &PrivateHandler{
		service: s,
	}
}

// Liveness checks the liveness of the server
func (h *PrivateHandler) Liveness(ctx context.Context, in *controllerPB.LivenessRequest) (*controllerPB.LivenessResponse, error) {
	return &controllerPB.LivenessResponse{
		HealthCheckResponse: &healthcheckPB.HealthCheckResponse{
			Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_NOT_SERVING,
		},
	}, nil

}

// Readiness checks the readiness of the server
func (h *PrivateHandler) Readiness(ctx context.Context, in *controllerPB.ReadinessRequest) (*controllerPB.ReadinessResponse, error) {
	return &controllerPB.ReadinessResponse{
		HealthCheckResponse: &healthcheckPB.HealthCheckResponse{
			Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_SERVING,
		},
	}, nil
}

func (h *PrivateHandler) GetResource(ctx context.Context, req *controllerPB.GetResourceRequest) (*controllerPB.GetResourceResponse, error) {
	resourceName, _, _ := strings.Cut(req.Name, "/watch")
	resource, err := h.service.GetResourceState(resourceName)

	if err != nil {
		return nil, err
	}

	return &controllerPB.GetResourceResponse{
		Resource: &controllerPB.Resource{
			Name:     resource.Name,
			State:    resource.State,
			Progress: resource.Progress,
		},
	}, nil
}

func (h *PrivateHandler) UpdateResource(ctx context.Context, req *controllerPB.UpdateResourceRequest) (*controllerPB.UpdateResourceResponse, error) {
	err := h.service.UpdateResourceState(req.Resource.Name, req.Resource.State)

	if err != nil {
		return nil, err
	}

	return &controllerPB.UpdateResourceResponse{
		Resource: req.Resource,
	}, nil
}

func (h *PrivateHandler) DeleteResource(ctx context.Context, req *controllerPB.DeleteResourceRequest) (*controllerPB.DeleteResourceResponse, error) {
	err := h.service.DeleteResourceState(req.Name)

	if err != nil {
		return nil, err
	}

	return &controllerPB.DeleteResourceResponse{}, nil
}
