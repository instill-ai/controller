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
	"github.com/instill-ai/controller/internal/triton"
	"github.com/instill-ai/controller/internal/util"

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
	GetResourceWorkflowId(resourceName string) (*string, error)
	UpdateResourceWorkflowId(resourceName string, workflowId string) error
	DeleteResourceWorkflowId(resourceName string) error
	getOperationInfo(workflowId string, resourceType string) (*longrunningpb.Operation, error)
	ProbeBackend() error
	ProbeModels() error
	ProbeSourceConnectors() error
	ProbeDestinationConnectors() error
	ProbePipelines() error
}

type service struct {
	etcdClient             etcdv3.Client
	tritonClient           inferenceserver.GRPCInferenceServiceClient
	mgmtPublicClient       mgmtPB.MgmtPublicServiceClient
	modelPublicClient      modelPB.ModelPublicServiceClient
	modelPrivateClient     modelPB.ModelPrivateServiceClient
	pipelinePublicClient   pipelinePB.PipelinePublicServiceClient
	pipelinePrivateClient  pipelinePB.PipelinePrivateServiceClient
	connectorPublicClient  connectorPB.ConnectorPublicServiceClient
	connectorPrivateClient connectorPB.ConnectorPrivateServiceClient
}

func NewService(
	e etcdv3.Client,
	t inferenceserver.GRPCInferenceServiceClient,
	mg mgmtPB.MgmtPublicServiceClient,
	mp modelPB.ModelPublicServiceClient,
	m modelPB.ModelPrivateServiceClient,
	p pipelinePB.PipelinePublicServiceClient,
	pp pipelinePB.PipelinePrivateServiceClient,
	c connectorPB.ConnectorPublicServiceClient,
	cp connectorPB.ConnectorPrivateServiceClient) Service {
	return &service{
		etcdClient:             e,
		tritonClient:           t,
		mgmtPublicClient:       mg,
		modelPublicClient:      mp,
		modelPrivateClient:     m,
		pipelinePublicClient:   p,
		pipelinePrivateClient:  pp,
		connectorPublicClient:  c,
		connectorPrivateClient: cp,
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
	case util.RESOURCE_TYPE_SOURCE_CONNECTOR, util.RESOURCE_TYPE_DESTINATION_CONNECTOR:
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
		return nil, fmt.Errorf(fmt.Sprintf("get resource type %s not implemented", resourceType))
	}
}

func (s *service) UpdateResourceState(resource *controllerPB.Resource) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Etcd.Timeout*time.Second)
	defer cancel()

	workflowId, _ := s.GetResourceWorkflowId(resource.Name)

	resourceType := strings.SplitN(resource.Name, "/", 4)[3]

	state := 0

	switch resourceType {
	case util.RESOURCE_TYPE_MODEL:
		state = int(resource.GetModelInstanceState())
	case util.RESOURCE_TYPE_PIPELINE:
		state = int(resource.GetPipelineState())
	case util.RESOURCE_TYPE_SOURCE_CONNECTOR, util.RESOURCE_TYPE_DESTINATION_CONNECTOR:
		state = int(resource.GetConnectorState())
	case util.RESOURCE_TYPE_SERVICE:
		state = int(resource.GetBackendState())
	default:
		return fmt.Errorf(fmt.Sprintf("update resource type %s not implemented", resourceType))
	}

	// only for models
	if workflowId != nil {
		opInfo, err := s.getOperationInfo(*workflowId, resourceType)

		if err != nil {
			return err
		}

		if !opInfo.Done {
			switch resourceType {
			case util.RESOURCE_TYPE_MODEL:
				state = int(modelPB.ModelInstance_STATE_UNSPECIFIED)
			case util.RESOURCE_TYPE_PIPELINE:
				state = int(pipelinePB.Pipeline_STATE_UNSPECIFIED)
			case util.RESOURCE_TYPE_SOURCE_CONNECTOR, util.RESOURCE_TYPE_DESTINATION_CONNECTOR:
				state = int(connectorPB.Connector_STATE_UNSPECIFIED)
			case util.RESOURCE_TYPE_SERVICE:
				state = int(healthcheckv1alpha.HealthCheckResponse_SERVING_STATUS_UNSPECIFIED)
			default:
				return fmt.Errorf(fmt.Sprintf("resource type %s not implemented", resourceType))
			}
		} else {
			if err := s.DeleteResourceWorkflowId(resource.Name); err != nil {
				return err
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

func (s *service) GetResourceWorkflowId(resourceName string) (*string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Etcd.Timeout*time.Second)
	defer cancel()

	resourceWorkflowId := util.ConvertWorkflfowToWorkflowResourceName(resourceName)

	resp, err := s.etcdClient.Get(ctx, resourceWorkflowId)

	if err != nil {
		return nil, err
	}

	kvs := resp.Kvs

	if len(kvs) == 0 {
		return nil, fmt.Errorf("workflowId not found in etcd storage")
	}

	workflowId := string(kvs[0].Value[:])

	return &workflowId, nil
}

func (s *service) UpdateResourceWorkflowId(resourceName string, workflowId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Etcd.Timeout*time.Second)
	defer cancel()

	resourceWorkflowId := util.ConvertWorkflfowToWorkflowResourceName(resourceName)

	_, err := s.etcdClient.Put(ctx, resourceWorkflowId, workflowId)

	if err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteResourceWorkflowId(resourceName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Etcd.Timeout*time.Second)
	defer cancel()

	resourceWorkflowId := util.ConvertWorkflfowToWorkflowResourceName(resourceName)

	_, err := s.etcdClient.Delete(ctx, resourceWorkflowId)

	if err != nil {
		return err
	}

	return nil
}

func (s *service) ProbeBackend() error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Server.Timeout*time.Second)
	defer cancel()

	logger, _ := logger.GetZapLogger()

	healthcheck := healthcheckv1alpha.HealthCheckResponse{
		Status: healthcheckv1alpha.HealthCheckResponse_SERVING_STATUS_UNSPECIFIED,
	}

	var backenServices = [...]string{
		config.Config.TritonServer.Host,
		config.Config.ConnectorBackend.Host,
		config.Config.ModelBackend.Host,
		config.Config.PipelineBackend.Host,
		config.Config.MgmtBackend.Host,
	}

	for _, hostname := range backenServices {
		switch hostname {
		case config.Config.TritonServer.Host:
			resp, err := s.tritonClient.ServerLive(ctx, &inferenceserver.ServerLiveRequest{})

			if err != nil {
				return err
			}
			if resp.GetLive() {
				healthcheck = healthcheckv1alpha.HealthCheckResponse{
					Status: healthcheckv1alpha.HealthCheckResponse_SERVING_STATUS_SERVING,
				}
			} else {
				healthcheck = healthcheckv1alpha.HealthCheckResponse{
					Status: healthcheckv1alpha.HealthCheckResponse_SERVING_STATUS_NOT_SERVING,
				}
			}
		case config.Config.ModelBackend.Host:
			resp, err := s.modelPublicClient.Liveness(ctx, &modelPB.LivenessRequest{})

			if err != nil {
				return err
			}
			healthcheck = *resp.GetHealthCheckResponse()
		case config.Config.PipelineBackend.Host:
			resp, err := s.pipelinePublicClient.Liveness(ctx, &pipelinePB.LivenessRequest{})

			if err != nil {
				return err
			}
			healthcheck = *resp.GetHealthCheckResponse()
		case config.Config.MgmtBackend.Host:
			resp, err := s.mgmtPublicClient.Liveness(ctx, &mgmtPB.LivenessRequest{})

			if err != nil {
				return err
			}
			healthcheck = *resp.GetHealthCheckResponse()
		case config.Config.ConnectorBackend.Host:
			resp, err := s.connectorPublicClient.Liveness(ctx, &connectorPB.LivenessRequest{})

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
			Name: util.ConvertServiceToResourceName(hostname),
			State: &controllerPB.Resource_BackendState{
				BackendState: state,
			},
		})

		if err != nil {
			return err
		}

		resp, _ := s.GetResourceState(util.ConvertServiceToResourceName(hostname))

		logger.Info(fmt.Sprintf("[Controller] Got %v", resp))
	}

	return nil
}

func (s *service) getOperationInfo(workflowId string, resourceType string) (*longrunningpb.Operation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Server.Timeout*time.Second)
	defer cancel()

	var operation *longrunningpb.Operation

	switch resourceType {
	case util.RESOURCE_TYPE_MODEL:
		op, err := s.modelPublicClient.GetModelOperation(ctx, &modelPB.GetModelOperationRequest{
			Name: fmt.Sprintf("operations/%s", workflowId),
		})
		if err != nil {
			return nil, err
		}
		operation = op.Operation
	case util.RESOURCE_TYPE_SOURCE_CONNECTOR, util.RESOURCE_TYPE_DESTINATION_CONNECTOR:
		op, err := s.connectorPublicClient.GetConnectorOperation(ctx, &connectorPB.GetConnectorOperationRequest{
			Name: fmt.Sprintf("operations/%s", workflowId),
		})
		if err != nil {
			return nil, err
		}
		operation = op.Operation
	}

	return operation, nil
}
