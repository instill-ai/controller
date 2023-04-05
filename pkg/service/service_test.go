package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/instill-ai/controller/pkg/service"
	"github.com/stretchr/testify/assert"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
	controllerPB "github.com/instill-ai/protogen-go/vdp/controller/v1alpha"
	healthcheckPB "github.com/instill-ai/protogen-go/vdp/healthcheck/v1alpha"
	modelPB "github.com/instill-ai/protogen-go/vdp/model/v1alpha"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

const serviceResourceName = "resources/name/types/services"
const modelResourceName = "resources/name/types/models"
const connectorResourceName = "resources/name/types/source-connectors"
const pipelineResourceName = "resources/name/types/pipelines"

type Client struct {
	etcdv3.Cluster
	etcdv3.KV
	etcdv3.Lease
	etcdv3.Watcher
	etcdv3.Auth
	etcdv3.Maintenance

	// Username is a user name for authentication.
	Username string
	// Password is a password for authentication.
	Password string
	// contains filtered or unexported fields
}

func TestGetResourceState(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t.Run("service", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockCluster := NewMockCluster(ctrl)
		mockKV := NewMockKV(ctrl)
		mockLease := NewMockLease(ctrl)
		mockWatcher := NewMockWatcher(ctrl)
		mockAuth := NewMockAuth(ctrl)
		mockMaintenance := NewMockMaintenance(ctrl)

		mockEtcdClient := etcdv3.Client{
			Cluster:     mockCluster,
			KV:          mockKV,
			Lease:       mockLease,
			Watcher:     mockWatcher,
			Auth:        mockAuth,
			Maintenance: mockMaintenance,
		}

		var resp *etcdv3.GetResponse

		mockKV.
			EXPECT().
			Get(ctx, serviceResourceName).
			Return(resp, nil).
			Times(1)

		s := service.NewService(mockEtcdClient, nil, nil, nil, nil, nil, nil, nil, nil)

		resource, err := s.GetResourceState(ctx, serviceResourceName)

		assert.Equal(t, healthcheckPB.HealthCheckResponse_SERVING_STATUS_UNSPECIFIED, resource.GetBackendState())

		assert.NoError(t, err)
	})
	t.Run("model", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockCluster := NewMockCluster(ctrl)
		mockKV := NewMockKV(ctrl)
		mockLease := NewMockLease(ctrl)
		mockWatcher := NewMockWatcher(ctrl)
		mockAuth := NewMockAuth(ctrl)
		mockMaintenance := NewMockMaintenance(ctrl)

		mockEtcdClient := etcdv3.Client{
			Cluster:     mockCluster,
			KV:          mockKV,
			Lease:       mockLease,
			Watcher:     mockWatcher,
			Auth:        mockAuth,
			Maintenance: mockMaintenance,
		}

		var resp *etcdv3.GetResponse

		mockKV.
			EXPECT().
			Get(ctx, modelResourceName).
			Return(resp, nil).
			Times(1)

		s := service.NewService(mockEtcdClient, nil, nil, nil, nil, nil, nil, nil, nil)

		resource, err := s.GetResourceState(ctx, modelResourceName)

		assert.Equal(t, modelPB.Model_STATE_UNSPECIFIED, resource.GetModelState())

		assert.NoError(t, err)
	})
	t.Run("connector", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockCluster := NewMockCluster(ctrl)
		mockKV := NewMockKV(ctrl)
		mockLease := NewMockLease(ctrl)
		mockWatcher := NewMockWatcher(ctrl)
		mockAuth := NewMockAuth(ctrl)
		mockMaintenance := NewMockMaintenance(ctrl)

		mockEtcdClient := etcdv3.Client{
			Cluster:     mockCluster,
			KV:          mockKV,
			Lease:       mockLease,
			Watcher:     mockWatcher,
			Auth:        mockAuth,
			Maintenance: mockMaintenance,
		}

		var resp *etcdv3.GetResponse

		mockKV.
			EXPECT().
			Get(ctx, connectorResourceName).
			Return(resp, nil).
			Times(1)

		s := service.NewService(mockEtcdClient, nil, nil, nil, nil, nil, nil, nil, nil)

		resource, err := s.GetResourceState(ctx, connectorResourceName)

		assert.Equal(t, connectorPB.Connector_STATE_UNSPECIFIED, resource.GetConnectorState())

		assert.NoError(t, err)
	})
	t.Run("pipeline", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockCluster := NewMockCluster(ctrl)
		mockKV := NewMockKV(ctrl)
		mockLease := NewMockLease(ctrl)
		mockWatcher := NewMockWatcher(ctrl)
		mockAuth := NewMockAuth(ctrl)
		mockMaintenance := NewMockMaintenance(ctrl)

		mockEtcdClient := etcdv3.Client{
			Cluster:     mockCluster,
			KV:          mockKV,
			Lease:       mockLease,
			Watcher:     mockWatcher,
			Auth:        mockAuth,
			Maintenance: mockMaintenance,
		}

		var resp *etcdv3.GetResponse

		mockKV.
			EXPECT().
			Get(ctx, pipelineResourceName).
			Return(resp, nil).
			Times(1)

		s := service.NewService(mockEtcdClient, nil, nil, nil, nil, nil, nil, nil, nil)

		resource, err := s.GetResourceState(ctx, pipelineResourceName)

		assert.Equal(t, pipelinePB.Pipeline_STATE_UNSPECIFIED, resource.GetPipelineState())

		assert.NoError(t, err)
	})
}

func TestUpdateResourceState(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t.Run("service", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockCluster := NewMockCluster(ctrl)
		mockKV := NewMockKV(ctrl)
		mockLease := NewMockLease(ctrl)
		mockWatcher := NewMockWatcher(ctrl)
		mockAuth := NewMockAuth(ctrl)
		mockMaintenance := NewMockMaintenance(ctrl)

		mockEtcdClient := etcdv3.Client{
			Cluster:     mockCluster,
			KV:          mockKV,
			Lease:       mockLease,
			Watcher:     mockWatcher,
			Auth:        mockAuth,
			Maintenance: mockMaintenance,
		}

		resource := controllerPB.Resource{
			Name: serviceResourceName,
			State: &controllerPB.Resource_BackendState{
				BackendState: healthcheckPB.HealthCheckResponse_SERVING_STATUS_UNSPECIFIED,
			},
		}

		mockKV.
			EXPECT().
			Get(ctx, fmt.Sprintf("%s/workflow", serviceResourceName)).
			Return(&etcdv3.GetResponse{}, nil).
			Times(1)

		mockKV.
			EXPECT().
			Put(ctx, serviceResourceName, string("0")).
			Return(&etcdv3.PutResponse{}, nil).
			Times(1)

		s := service.NewService(mockEtcdClient, nil, nil, nil, nil, nil, nil, nil, nil)

		err := s.UpdateResourceState(ctx, &resource)

		assert.NoError(t, err)
	})

	t.Run("model", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockCluster := NewMockCluster(ctrl)
		mockKV := NewMockKV(ctrl)
		mockLease := NewMockLease(ctrl)
		mockWatcher := NewMockWatcher(ctrl)
		mockAuth := NewMockAuth(ctrl)
		mockMaintenance := NewMockMaintenance(ctrl)

		mockEtcdClient := etcdv3.Client{
			Cluster:     mockCluster,
			KV:          mockKV,
			Lease:       mockLease,
			Watcher:     mockWatcher,
			Auth:        mockAuth,
			Maintenance: mockMaintenance,
		}

		resource := controllerPB.Resource{
			Name: modelResourceName,
			State: &controllerPB.Resource_ModelState{
				ModelState: modelPB.Model_STATE_UNSPECIFIED,
			},
		}

		mockKV.
			EXPECT().
			Get(ctx, fmt.Sprintf("%s/workflow", modelResourceName)).
			Return(&etcdv3.GetResponse{}, nil).
			Times(1)

		mockKV.
			EXPECT().
			Put(ctx, modelResourceName, string("0")).
			Return(&etcdv3.PutResponse{}, nil).
			Times(1)

		s := service.NewService(mockEtcdClient, nil, nil, nil, nil, nil, nil, nil, nil)

		err := s.UpdateResourceState(ctx, &resource)

		assert.NoError(t, err)
	})

	t.Run("connector", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockCluster := NewMockCluster(ctrl)
		mockKV := NewMockKV(ctrl)
		mockLease := NewMockLease(ctrl)
		mockWatcher := NewMockWatcher(ctrl)
		mockAuth := NewMockAuth(ctrl)
		mockMaintenance := NewMockMaintenance(ctrl)

		mockEtcdClient := etcdv3.Client{
			Cluster:     mockCluster,
			KV:          mockKV,
			Lease:       mockLease,
			Watcher:     mockWatcher,
			Auth:        mockAuth,
			Maintenance: mockMaintenance,
		}

		resource := controllerPB.Resource{
			Name: connectorResourceName,
			State: &controllerPB.Resource_ConnectorState{
				ConnectorState: connectorPB.Connector_STATE_UNSPECIFIED,
			},
		}

		mockKV.
			EXPECT().
			Get(ctx, fmt.Sprintf("%s/workflow", connectorResourceName)).
			Return(&etcdv3.GetResponse{}, nil).
			Times(1)

		mockKV.
			EXPECT().
			Put(ctx, connectorResourceName, string("0")).
			Return(&etcdv3.PutResponse{}, nil).
			Times(1)

		s := service.NewService(mockEtcdClient, nil, nil, nil, nil, nil, nil, nil, nil)

		err := s.UpdateResourceState(ctx, &resource)

		assert.NoError(t, err)
	})

	t.Run("pipeline", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockCluster := NewMockCluster(ctrl)
		mockKV := NewMockKV(ctrl)
		mockLease := NewMockLease(ctrl)
		mockWatcher := NewMockWatcher(ctrl)
		mockAuth := NewMockAuth(ctrl)
		mockMaintenance := NewMockMaintenance(ctrl)

		mockEtcdClient := etcdv3.Client{
			Cluster:     mockCluster,
			KV:          mockKV,
			Lease:       mockLease,
			Watcher:     mockWatcher,
			Auth:        mockAuth,
			Maintenance: mockMaintenance,
		}

		resource := controllerPB.Resource{
			Name: pipelineResourceName,
			State: &controllerPB.Resource_PipelineState{
				PipelineState: pipelinePB.Pipeline_STATE_UNSPECIFIED,
			},
		}

		mockKV.
			EXPECT().
			Get(ctx, fmt.Sprintf("%s/workflow", pipelineResourceName)).
			Return(&etcdv3.GetResponse{}, nil).
			Times(1)

		mockKV.
			EXPECT().
			Put(ctx, pipelineResourceName, string("0")).
			Return(&etcdv3.PutResponse{}, nil).
			Times(1)

		s := service.NewService(mockEtcdClient, nil, nil, nil, nil, nil, nil, nil, nil)

		err := s.UpdateResourceState(ctx, &resource)

		assert.NoError(t, err)
	})
}

func TestDeleteResourceState(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t.Run("service", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockCluster := NewMockCluster(ctrl)
		mockKV := NewMockKV(ctrl)
		mockLease := NewMockLease(ctrl)
		mockWatcher := NewMockWatcher(ctrl)
		mockAuth := NewMockAuth(ctrl)
		mockMaintenance := NewMockMaintenance(ctrl)

		mockEtcdClient := etcdv3.Client{
			Cluster:     mockCluster,
			KV:          mockKV,
			Lease:       mockLease,
			Watcher:     mockWatcher,
			Auth:        mockAuth,
			Maintenance: mockMaintenance,
		}

		var resp *etcdv3.DeleteResponse

		mockKV.
			EXPECT().
			Delete(ctx, serviceResourceName).
			Return(resp, nil).
			Times(1)

		s := service.NewService(mockEtcdClient, nil, nil, nil, nil, nil, nil, nil, nil)

		err := s.DeleteResourceState(ctx, serviceResourceName)

		assert.NoError(t, err)
	})
	t.Run("model", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockCluster := NewMockCluster(ctrl)
		mockKV := NewMockKV(ctrl)
		mockLease := NewMockLease(ctrl)
		mockWatcher := NewMockWatcher(ctrl)
		mockAuth := NewMockAuth(ctrl)
		mockMaintenance := NewMockMaintenance(ctrl)

		mockEtcdClient := etcdv3.Client{
			Cluster:     mockCluster,
			KV:          mockKV,
			Lease:       mockLease,
			Watcher:     mockWatcher,
			Auth:        mockAuth,
			Maintenance: mockMaintenance,
		}

		var resp *etcdv3.DeleteResponse

		mockKV.
			EXPECT().
			Delete(ctx, modelResourceName).
			Return(resp, nil).
			Times(1)

		s := service.NewService(mockEtcdClient, nil, nil, nil, nil, nil, nil, nil, nil)

		err := s.DeleteResourceState(ctx, modelResourceName)

		assert.NoError(t, err)
	})
	t.Run("connector", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockCluster := NewMockCluster(ctrl)
		mockKV := NewMockKV(ctrl)
		mockLease := NewMockLease(ctrl)
		mockWatcher := NewMockWatcher(ctrl)
		mockAuth := NewMockAuth(ctrl)
		mockMaintenance := NewMockMaintenance(ctrl)

		mockEtcdClient := etcdv3.Client{
			Cluster:     mockCluster,
			KV:          mockKV,
			Lease:       mockLease,
			Watcher:     mockWatcher,
			Auth:        mockAuth,
			Maintenance: mockMaintenance,
		}

		var resp *etcdv3.DeleteResponse

		mockKV.
			EXPECT().
			Delete(ctx, connectorResourceName).
			Return(resp, nil).
			Times(1)

		s := service.NewService(mockEtcdClient, nil, nil, nil, nil, nil, nil, nil, nil)

		err := s.DeleteResourceState(ctx, connectorResourceName)

		assert.NoError(t, err)
	})
	t.Run("pipeline", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockCluster := NewMockCluster(ctrl)
		mockKV := NewMockKV(ctrl)
		mockLease := NewMockLease(ctrl)
		mockWatcher := NewMockWatcher(ctrl)
		mockAuth := NewMockAuth(ctrl)
		mockMaintenance := NewMockMaintenance(ctrl)

		mockEtcdClient := etcdv3.Client{
			Cluster:     mockCluster,
			KV:          mockKV,
			Lease:       mockLease,
			Watcher:     mockWatcher,
			Auth:        mockAuth,
			Maintenance: mockMaintenance,
		}

		var resp *etcdv3.DeleteResponse

		mockKV.
			EXPECT().
			Delete(ctx, pipelineResourceName).
			Return(resp, nil).
			Times(1)

		s := service.NewService(mockEtcdClient, nil, nil, nil, nil, nil, nil, nil, nil)

		err := s.DeleteResourceState(ctx, pipelineResourceName)

		assert.NoError(t, err)
	})
}
