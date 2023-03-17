package service

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/instill-ai/controller/config"
	"github.com/instill-ai/controller/internal/logger"
	"github.com/instill-ai/controller/pkg/datamodel"
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
	GetResourceState(resourceName string) (*datamodel.Resource, error)
	UpdateResourceState(resourceName string, state controllerPB.Resource_State) error
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

func (s *service) GetResourceState(resourceName string) (*datamodel.Resource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Etcd.Timeout*time.Second)
	defer cancel()

	// logger, _ := logger.GetZapLogger()

	resp, err := s.etcdClient.Get(ctx, resourceName)

	if err != nil {
		return nil, err
	}

	kvs := resp.Kvs

	if len(kvs) == 0 {
		return nil, fmt.Errorf("resource not found in etcd storage")
	}

	resource := datamodel.Resource{
		Name:  resourceName,
		State: controllerPB.Resource_State(kvs[0].Value[0]),
	}

	// logger.Info(fmt.Sprintf("[Etcd Client] Get %v", resource))

	return &resource, nil
}

func (s *service) UpdateResourceState(resourceName string, state controllerPB.Resource_State) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Etcd.Timeout*time.Second)
	defer cancel()

	logger, _ := logger.GetZapLogger()

	workflowID, _ := s.GetResourceWorkflowID(resourceName)

	if workflowID != nil {
		opInfo, err := s.GetOperationInfo(*workflowID)

		if err != nil {
			return err
		}

		if !opInfo.Done {
			state = controllerPB.Resource_STATE_UNSPECIFIED
		}
	}

	_, err := s.etcdClient.Put(ctx, resourceName, string(state))

	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("[Etcd Client] Got %v with %v", resourceName, state))

	return nil
}

func (s *service) DeleteResourceState(resourceName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Etcd.Timeout*time.Second)
	defer cancel()

	logger, _ := logger.GetZapLogger()

	_, err := s.etcdClient.Delete(ctx, resourceName)

	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("[Etcd Client] Delete %v", resourceName))

	return nil
}

func (s *service) GetResourceWorkflowID(resourceName string) (*string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Etcd.Timeout*time.Second)
	defer cancel()

	// logger, _ := logger.GetZapLogger()

	resourceName = "workflow/" + resourceName

	resp, err := s.etcdClient.Get(ctx, resourceName)

	if err != nil {
		return nil, err
	}

	kvs := resp.Kvs

	if len(kvs) == 0 {
		return nil, fmt.Errorf("workflowID not found in etcd storage")
	}

	workflowId := string(kvs[0].Value[:])

	// logger.Info(fmt.Sprintf("[Etcd Client] Get %v, id: %v", resourceName, workflowId))

	return &workflowId, nil
}

func (s *service) UpdateResourceWorkflowID(resourceName string, workflowID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Etcd.Timeout*time.Second)
	defer cancel()

	// logger, _ := logger.GetZapLogger()

	resourceName = "workflow/" + resourceName

	_, err := s.etcdClient.Put(ctx, resourceName, workflowID)

	if err != nil {
		return err
	}

	// logger.Info(fmt.Sprintf("[Etcd Client] Update %v with %v", resourceName, workflowID))

	return nil
}

func (s *service) DeleteResourceWorkflowID(resourceName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Etcd.Timeout*time.Second)
	defer cancel()

	// logger, _ := logger.GetZapLogger()

	resourceName = "workflow/" + resourceName

	_, err := s.etcdClient.Delete(ctx, resourceName)

	if err != nil {
		return err
	}

	// logger.Info(fmt.Sprintf("[Etcd Client] Delete %v", resourceName))

	return nil
}

func (s *service) ProbeBackend(serviceName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Server.Timeout*time.Second)
	defer cancel()

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

	state := controllerPB.Resource_STATE_ERROR
	switch healthcheck.Status {
	case 0:
		state = controllerPB.Resource_STATE_UNSPECIFIED
	case 1:
		state = controllerPB.Resource_STATE_ONLINE
	case 2:
		state = controllerPB.Resource_STATE_OFFLINE
	}
	err := s.UpdateResourceState(serviceName, state)

	if err != nil {
		return err
	}
	return nil
}

func (s *service) ProbeModels() error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Server.Timeout*time.Second)
	defer cancel()

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

		err = s.UpdateResourceState(resp.Resource.Name, resp.Resource.State)

		if err != nil {
			return err
		}
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
