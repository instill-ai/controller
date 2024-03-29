package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"

	etcdv3 "go.etcd.io/etcd/client/v3"

	"github.com/instill-ai/controller/config"
	"github.com/instill-ai/controller/internal/triton"
	"github.com/instill-ai/controller/internal/util"
	"github.com/instill-ai/controller/pkg/logger"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
	controllerPB "github.com/instill-ai/protogen-go/vdp/controller/v1alpha"
	healthcheckPB "github.com/instill-ai/protogen-go/vdp/healthcheck/v1alpha"
	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
	modelPB "github.com/instill-ai/protogen-go/vdp/model/v1alpha"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
)

type Service interface {
	GetResourceState(ctx context.Context, resourcePermalink string) (*controllerPB.Resource, error)
	UpdateResourceState(ctx context.Context, resource *controllerPB.Resource) error
	DeleteResourceState(ctx context.Context, resourcePermalink string) error
	GetResourceWorkflowId(ctx context.Context, resourcePermalink string) (*string, error)
	UpdateResourceWorkflowId(ctx context.Context, resourcePermalink string, workflowId string) error
	DeleteResourceWorkflowId(ctx context.Context, resourcePermalink string) error
	ProbeBackend(ctx context.Context, cancel context.CancelFunc) error
	ProbeModels(ctx context.Context, cancel context.CancelFunc) error
	ProbeSourceConnectors(ctx context.Context, cancel context.CancelFunc) error
	ProbeDestinationConnectors(ctx context.Context, cancel context.CancelFunc) error
	ProbePipelines(ctx context.Context, cancel context.CancelFunc) error
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

func (s *service) GetResourceState(ctx context.Context, resourcePermalink string) (*controllerPB.Resource, error) {
	resp, err := s.etcdClient.Get(ctx, resourcePermalink)

	if err != nil {
		return nil, err
	}

	kvs := resp.Kvs

	if len(kvs) == 0 {
		return nil, fmt.Errorf(fmt.Sprintf("resource %v not found in etcd storage", resourcePermalink))
	}

	resourceType := strings.SplitN(resourcePermalink, "/", 4)[3]

	stateEnumValue, _ := strconv.ParseInt(string(kvs[0].Value[:]), 10, 32)

	switch resourceType {
	case util.RESOURCE_TYPE_MODEL:
		return &controllerPB.Resource{
			ResourcePermalink: resourcePermalink,
			State: &controllerPB.Resource_ModelState{
				ModelState: modelPB.Model_State(stateEnumValue),
			},
			Progress: nil,
		}, nil
	case util.RESOURCE_TYPE_PIPELINE:
		return &controllerPB.Resource{
			ResourcePermalink: resourcePermalink,
			State: &controllerPB.Resource_PipelineState{
				PipelineState: pipelinePB.Pipeline_State(stateEnumValue),
			},
			Progress: nil,
		}, nil
	case util.RESOURCE_TYPE_SOURCE_CONNECTOR, util.RESOURCE_TYPE_DESTINATION_CONNECTOR:
		return &controllerPB.Resource{
			ResourcePermalink: resourcePermalink,
			State: &controllerPB.Resource_ConnectorState{
				ConnectorState: connectorPB.Connector_State(stateEnumValue),
			},
			Progress: nil,
		}, nil
	case util.RESOURCE_TYPE_SERVICE:
		return &controllerPB.Resource{
			ResourcePermalink: resourcePermalink,
			State: &controllerPB.Resource_BackendState{
				BackendState: healthcheckPB.HealthCheckResponse_ServingStatus(stateEnumValue),
			},
		}, nil
	default:
		return nil, fmt.Errorf(fmt.Sprintf("get resource type %s not implemented", resourceType))
	}
}

func (s *service) UpdateResourceState(ctx context.Context, resource *controllerPB.Resource) error {
	resourceType := strings.SplitN(resource.ResourcePermalink, "/", 4)[3]

	state := 0

	switch resourceType {
	case util.RESOURCE_TYPE_MODEL:
		state = int(resource.GetModelState())
	case util.RESOURCE_TYPE_PIPELINE:
		state = int(resource.GetPipelineState())
	case util.RESOURCE_TYPE_SOURCE_CONNECTOR, util.RESOURCE_TYPE_DESTINATION_CONNECTOR:
		state = int(resource.GetConnectorState())
	case util.RESOURCE_TYPE_SERVICE:
		state = int(resource.GetBackendState())
	default:
		return fmt.Errorf(fmt.Sprintf("update resource type %s not implemented", resourceType))
	}

	if _, err := s.etcdClient.Put(ctx, resource.ResourcePermalink, fmt.Sprint(state)); err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteResourceState(ctx context.Context, resourcePermalink string) error {
	_, err := s.etcdClient.Delete(ctx, resourcePermalink)

	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetResourceWorkflowId(ctx context.Context, resourcePermalink string) (*string, error) {
	resourceWorkflowId := util.ConvertResourcePermalinkToWorkflowName(resourcePermalink)

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

func (s *service) UpdateResourceWorkflowId(ctx context.Context, resourcePermalink string, workflowId string) error {
	resourceWorkflowId := util.ConvertResourcePermalinkToWorkflowName(resourcePermalink)

	_, err := s.etcdClient.Put(ctx, resourceWorkflowId, workflowId)

	if err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteResourceWorkflowId(ctx context.Context, resourcePermalink string) error {
	resourceWorkflowId := util.ConvertResourcePermalinkToWorkflowName(resourcePermalink)

	_, err := s.etcdClient.Delete(ctx, resourceWorkflowId)

	if err != nil {
		return err
	}

	return nil
}

func (s *service) ProbeBackend(ctx context.Context, cancel context.CancelFunc) error {
	defer cancel()

	logger, _ := logger.GetZapLogger(ctx)

	var wg sync.WaitGroup

	healthcheck := healthcheckPB.HealthCheckResponse{
		Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_UNSPECIFIED,
	}

	var backenServices = [...]string{
		config.Config.TritonServer.Host,
		config.Config.ConnectorBackend.Host,
		config.Config.ModelBackend.Host,
		config.Config.PipelineBackend.Host,
		config.Config.MgmtBackend.Host,
	}

	wg.Add(len(backenServices))

	for _, hostname := range backenServices {
		go func(hostname string) {
			defer wg.Done()

			switch hostname {
			case config.Config.TritonServer.Host:
				resp, err := s.tritonClient.ServerLive(ctx, &inferenceserver.ServerLiveRequest{})

				if err != nil {
					healthcheck = healthcheckPB.HealthCheckResponse{
						Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_NOT_SERVING,
					}
				} else {
					if resp.GetLive() {
						healthcheck = healthcheckPB.HealthCheckResponse{
							Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_SERVING,
						}
					} else {
						healthcheck = healthcheckPB.HealthCheckResponse{
							Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_NOT_SERVING,
						}
					}
				}
			case config.Config.ModelBackend.Host:
				resp, err := s.modelPublicClient.Liveness(ctx, &modelPB.LivenessRequest{})

				if err != nil {
					healthcheck = healthcheckPB.HealthCheckResponse{
						Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_NOT_SERVING,
					}
				} else {
					healthcheck = *resp.GetHealthCheckResponse()
				}
			case config.Config.PipelineBackend.Host:
				resp, err := s.pipelinePublicClient.Liveness(ctx, &pipelinePB.LivenessRequest{})

				if err != nil {
					healthcheck = healthcheckPB.HealthCheckResponse{
						Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_NOT_SERVING,
					}
				} else {
					healthcheck = *resp.GetHealthCheckResponse()
				}
			case config.Config.MgmtBackend.Host:
				resp, err := s.mgmtPublicClient.Liveness(ctx, &mgmtPB.LivenessRequest{})

				if err != nil {
					healthcheck = healthcheckPB.HealthCheckResponse{
						Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_NOT_SERVING,
					}
				} else {
					healthcheck = *resp.GetHealthCheckResponse()
				}
			case config.Config.ConnectorBackend.Host:
				resp, err := s.connectorPublicClient.Liveness(ctx, &connectorPB.LivenessRequest{})

				if err != nil {
					healthcheck = healthcheckPB.HealthCheckResponse{
						Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_NOT_SERVING,
					}
				} else {
					healthcheck = *resp.GetHealthCheckResponse()
				}
			}

			err := s.UpdateResourceState(ctx, &controllerPB.Resource{
				ResourcePermalink: util.ConvertServiceToResourceName(hostname),
				State: &controllerPB.Resource_BackendState{
					BackendState: healthcheck.Status,
				},
			})

			if err != nil {
				logger.Error(err.Error())
				return
			}

			resp, _ := s.GetResourceState(ctx, util.ConvertServiceToResourceName(hostname))

			logger.Info(fmt.Sprintf("[Controller] Got %v", resp))
		}(hostname)
	}

	wg.Wait()

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
	}

	return operation, nil
}
