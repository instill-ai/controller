package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/instill-ai/controller/config"
	"github.com/instill-ai/controller/internal/logger"
	"github.com/instill-ai/controller/internal/util"
	"github.com/instill-ai/model-backend/pkg/repository"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
	controllerPB "github.com/instill-ai/protogen-go/vdp/controller/v1alpha"
	healthcheckv1alpha "github.com/instill-ai/protogen-go/vdp/healthcheck/v1alpha"
	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
	modelPB "github.com/instill-ai/protogen-go/vdp/model/v1alpha"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

type Service interface {
	GetResourceState(resourceName string) (*controllerPB.Resource, error)
	UpdateResourceState(resource *controllerPB.Resource) error
	DeleteResourceState(resourceName string) error
	GetResourceWorkflowID(resourceName string) (*string, error)
	UpdateResourceWorkflowID(resourceName string, workflowID string) error
	DeleteResourceWorkflowID(resourceName string) error
	GetOperationInfo(workflowID string) (*longrunningpb.Operation, error)
	ProbeBackend(serviceName string) error
	ProbeModels() error
}

type service struct {
	etcdClient         etcdv3.Client
	modelPublicClient  modelPB.ModelPublicServiceClient
	modelPrivateClient modelPB.ModelPrivateServiceClient
	mgmtClient         mgmtPB.MgmtPublicServiceClient
	pipelineClient     pipelinePB.PipelinePublicServiceClient
	connectorClient    connectorPB.ConnectorPublicServiceClient
}

func NewService(
	e etcdv3.Client,
	m modelPB.ModelPrivateServiceClient,
	mp modelPB.ModelPublicServiceClient,
	mg mgmtPB.MgmtPublicServiceClient,
	p pipelinePB.PipelinePublicServiceClient,
	c connectorPB.ConnectorPublicServiceClient) Service {
	return &service{
		etcdClient:         e,
		modelPublicClient:  mp,
		modelPrivateClient: m,
		mgmtClient:         mg,
		pipelineClient:     p,
		connectorClient:    c,
	}
}

func (s *service) GetResourceState(resourceName string) (*controllerPB.Resource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Etcd.Timeout*time.Second)
	defer cancel()

	resp, err := s.etcdClient.Get(ctx, resourceName)

	if err != nil {
		return nil, err
	}

	kvs := resp.Kvs

	if len(kvs) == 0 {
		return nil, fmt.Errorf(fmt.Sprintf("resource %v not found in etcd storage ", resourceName))
	}

	resourceType := strings.SplitN(resourceName, "/", 4)[3]

	stateEnumValue, _ := strconv.ParseInt(string(kvs[0].Value[:]), 10, 32)

	switch resourceType {
	case util.RESOURCE_TYPE_MODEL:
		return &controllerPB.Resource{
			Name: resourceName,
			State: &controllerPB.Resource_ModelInstanceState{
				ModelInstanceState: modelPB.ModelInstance_State(stateEnumValue),
			},
			Progress: nil,
		}, nil
	case util.RESOURCE_TYPE_PIPELINE:
		return &controllerPB.Resource{
			Name: resourceName,
			State: &controllerPB.Resource_PipelineState{
				PipelineState: pipelinePB.Pipeline_State(stateEnumValue),
			},
			Progress: nil,
		}, nil
	case util.RESOURCE_TYPE_CONNECTOR:
		return &controllerPB.Resource{
			Name: resourceName,
			State: &controllerPB.Resource_ConnectorState{
				ConnectorState: connectorPB.Connector_State(stateEnumValue),
			},
			Progress: nil,
		}, nil
	case util.RESOURCE_TYPE_SERVICE:
		return &controllerPB.Resource{
			Name: resourceName,
			State: &controllerPB.Resource_BackendState{
				BackendState: healthcheckv1alpha.HealthCheckResponse_ServingStatus(stateEnumValue),
			},
		}, nil
	default:
		return nil, fmt.Errorf("resource type not implemented")
	}
}

func (s *service) UpdateResourceState(resource *controllerPB.Resource) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Etcd.Timeout*time.Second)
	defer cancel()

	workflowID, _ := s.GetResourceWorkflowID(resource.Name)

	resourceType := strings.SplitN(resource.Name, "/", 4)[3]

	state := 0

	switch resourceType {
	case util.RESOURCE_TYPE_MODEL:
		state = int(resource.GetModelInstanceState())
	case util.RESOURCE_TYPE_PIPELINE:
		state = int(resource.GetPipelineState())
	case util.RESOURCE_TYPE_CONNECTOR:
		state = int(resource.GetConnectorState())
	case util.RESOURCE_TYPE_SERVICE:
		state = int(resource.GetBackendState())
	default:
		return fmt.Errorf("resource type not implemented")
	}

	if workflowID != nil {
		opInfo, err := s.GetOperationInfo(*workflowID)

		if err != nil {
			return err
		}

		if !opInfo.Done {
			switch resourceType {
			case util.RESOURCE_TYPE_MODEL:
				state = int(modelPB.ModelInstance_STATE_UNSPECIFIED)
			case util.RESOURCE_TYPE_PIPELINE:
				state = int(pipelinePB.Pipeline_STATE_UNSPECIFIED)
			case util.RESOURCE_TYPE_CONNECTOR:
				state = int(connectorPB.Connector_STATE_UNSPECIFIED)
			case util.RESOURCE_TYPE_SERVICE:
				state = int(healthcheckv1alpha.HealthCheckResponse_SERVING_STATUS_UNSPECIFIED)
			default:
				return fmt.Errorf("resource type not implemented")
			}
		}
	}

	_, err := s.etcdClient.Put(ctx, resource.Name, fmt.Sprint(state))

	if err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteResourceState(resourceName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Etcd.Timeout*time.Second)
	defer cancel()

	_, err := s.etcdClient.Delete(ctx, resourceName)

	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetResourceWorkflowID(resourceName string) (*string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Etcd.Timeout*time.Second)
	defer cancel()

	resourceWorkflowId := util.ConvertWorkflfowToResourceWorkflow(resourceName)

	resp, err := s.etcdClient.Get(ctx, resourceWorkflowId)

	if err != nil {
		return nil, err
	}

	kvs := resp.Kvs

	if len(kvs) == 0 {
		return nil, fmt.Errorf("workflowID not found in etcd storage")
	}

	workflowId := string(kvs[0].Value[:])

	return &workflowId, nil
}

func (s *service) UpdateResourceWorkflowID(resourceName string, workflowID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Etcd.Timeout*time.Second)
	defer cancel()

	resourceWorkflowId := util.ConvertWorkflfowToResourceWorkflow(resourceName)

	_, err := s.etcdClient.Put(ctx, resourceWorkflowId, workflowID)

	if err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteResourceWorkflowID(resourceName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Etcd.Timeout*time.Second)
	defer cancel()

	resourceWorkflowId := util.ConvertWorkflfowToResourceWorkflow(resourceName)

	_, err := s.etcdClient.Delete(ctx, resourceWorkflowId)

	if err != nil {
		return err
	}

	return nil
}

func (s *service) ProbeBackend(serviceName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Server.Timeout*time.Second)
	defer cancel()

	logger, _ := logger.GetZapLogger()

	healthcheck := healthcheckv1alpha.HealthCheckResponse{}
	switch serviceName {
	case config.Config.ModelBackend.Host:
		resp, err := s.modelPublicClient.Liveness(ctx, &modelPB.LivenessRequest{})

		if err != nil {
			return err
		}
		healthcheck = *resp.GetHealthCheckResponse()
	case config.Config.PipelineBackend.Host:
		resp, err := s.pipelineClient.Liveness(ctx, &pipelinePB.LivenessRequest{})

		if err != nil {
			return err
		}
		healthcheck = *resp.GetHealthCheckResponse()
	case config.Config.MgmtBackend.Host:
		resp, err := s.mgmtClient.Liveness(ctx, &mgmtPB.LivenessRequest{})

		if err != nil {
			return err
		}
		healthcheck = *resp.GetHealthCheckResponse()
	case config.Config.ConnectorBackend.Host:
		resp, err := s.connectorClient.Liveness(ctx, &connectorPB.LivenessRequest{})

		if err != nil {
			return err
		}
		healthcheck = *resp.GetHealthCheckResponse()
	}

	state := healthcheck.Status
	switch healthcheck.Status {
	case 0:
		state = healthcheckv1alpha.HealthCheckResponse_SERVING_STATUS_UNSPECIFIED
	case 1:
		state = healthcheckv1alpha.HealthCheckResponse_SERVING_STATUS_SERVING
	case 2:
		state = healthcheckv1alpha.HealthCheckResponse_SERVING_STATUS_NOT_SERVING
	}

	err := s.UpdateResourceState(&controllerPB.Resource{
		Name: util.ConvertServiceToResourceName(serviceName),
		State: &controllerPB.Resource_BackendState{
			BackendState: state,
		},
	})

	if err != nil {
		return err
	}

	resp, _ := s.GetResourceState(util.ConvertServiceToResourceName(serviceName))

	logger.Info(fmt.Sprintf("[Controller] Got %v", resp))

	return nil
}

func (s *service) ProbeModels() error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Server.Timeout*time.Second)
	defer cancel()

	logger, _ := logger.GetZapLogger()

	resp, err := s.modelPublicClient.ListModels(ctx, &modelPB.ListModelsRequest{})

	if err != nil {
		return err
	}

	models := resp.Models
	nextPageToken := &resp.NextPageToken
	totalSize := resp.TotalSize

	for totalSize > repository.DefaultPageSize {
		resp, err := s.modelPublicClient.ListModels(ctx, &modelPB.ListModelsRequest{
			PageToken: nextPageToken,
		})

		if err != nil {
			return err
		}

		nextPageToken = &resp.NextPageToken
		totalSize -= repository.DefaultPageSize
		models = append(models, resp.Models...)
	}

	modelInstances := []*modelPB.ModelInstance{}
	for _, model := range models {
		view := modelPB.View_VIEW_FULL
		resp, err := s.modelPublicClient.ListModelInstances(ctx, &modelPB.ListModelInstancesRequest{
			Parent: model.Name,
			View:   &view,
		})
		if err != nil {
			return err
		}

		nextPageToken := &resp.NextPageToken
		totalSize := resp.TotalSize
		modelInstances = append(modelInstances, resp.Instances...)

		for totalSize > repository.DefaultPageSize {
			resp, err := s.modelPublicClient.ListModelInstances(ctx, &modelPB.ListModelInstancesRequest{
				Parent:    model.Name,
				PageToken: nextPageToken,
				View:      &view,
			})

			if err != nil {
				return err
			}

			nextPageToken = &resp.NextPageToken
			totalSize -= repository.DefaultPageSize
			modelInstances = append(modelInstances, resp.Instances...)
		}
	}

	for _, modelInstance := range modelInstances {
		resp, err := s.modelPrivateClient.CheckModelInstance(ctx, &modelPB.CheckModelInstanceRequest{
			Name: modelInstance.Name,
		})

		if err != nil {
			return err
		}

		err = s.UpdateResourceState(&controllerPB.Resource{
			Name: util.ConvertModelToResourceName(modelInstance.Name),
			State: &controllerPB.Resource_ModelInstanceState{
				ModelInstanceState: resp.State,
			},
		})

		if err != nil {
			return err
		}

		logResp, _ := s.GetResourceState(util.ConvertModelToResourceName(modelInstance.Name))
		logger.Info(fmt.Sprintf("[Controller] Got %v", logResp))
	}

	return nil
}

func (s *service) GetOperationInfo(workflowID string) (*longrunningpb.Operation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Server.Timeout*time.Second)
	defer cancel()

	operation, err := s.modelPublicClient.GetModelOperation(ctx, &modelPB.GetModelOperationRequest{
		Name: fmt.Sprintf("operations/%s", workflowID),
	})

	if err != nil {
		return nil, err
	}

	return operation.Operation, nil
}
