package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/instill-ai/controller/config"
	"github.com/instill-ai/controller/pkg/datamodel"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
	controllerPB "github.com/instill-ai/protogen-go/vdp/controller/v1alpha"
	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
	modelPB "github.com/instill-ai/protogen-go/vdp/model/v1alpha"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

type Service interface {
	GetResourceState(resourceName string) (*datamodel.Resource, error)
	UpdateResourceState(resourceName string, state controllerPB.Resource_State) error
	DeleteResourceState(resourceName string) error
}

type service struct {
	etcdClient      etcdv3.Client
	modelClient     modelPB.ModelPrivateServiceClient
	mgmtClient      mgmtPB.MgmtPublicServiceClient
	pipelineClient  pipelinePB.PipelinePublicServiceClient
	connectorClient connectorPB.ConnectorPublicServiceClient
}

func NewService(e etcdv3.Client, m modelPB.ModelPrivateServiceClient, mg mgmtPB.MgmtPublicServiceClient, p pipelinePB.PipelinePublicServiceClient, c connectorPB.ConnectorPublicServiceClient) Service {
	return &service{
		etcdClient:  e,
		modelClient: m,
		mgmtClient: mg,
		pipelineClient: p,
		connectorClient: c,
	}
}

func (s *service) GetResourceState(resourceName string) (*datamodel.Resource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Etcd.Timeout*time.Second)
	defer cancel()

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

	return &resource, nil
}

func (s *service) UpdateResourceState(resourceName string, state controllerPB.Resource_State) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Etcd.Timeout*time.Second)
	defer cancel()
	fmt.Println("--------------------------------PUT: ", resourceName, state)

	fmt.Println("--------------------------------PUT: ", strconv.FormatInt(int64(state), 16), "num:", state.Number())
	_, err := s.etcdClient.Put(ctx, resourceName, string(state))

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
