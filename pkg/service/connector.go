package service

import (
	"context"
	"fmt"
	"strconv"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/instill-ai/controller/internal/logger"
	"github.com/instill-ai/controller/internal/util"
	"github.com/instill-ai/model-backend/pkg/repository"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
	controllerPB "github.com/instill-ai/protogen-go/vdp/controller/v1alpha"
)

func (s *service) ProbeSourceConnectors(ctx context.Context, cancel context.CancelFunc) error {
	defer cancel()

	resp, err := s.connectorPublicClient.ListSourceConnectors(ctx, &connectorPB.ListSourceConnectorsRequest{})

	if err != nil {
		return err
	}

	connectors := resp.SourceConnectors
	nextPageToken := &resp.NextPageToken
	totalSize := resp.TotalSize

	for totalSize > repository.DefaultPageSize {
		resp, err := s.connectorPublicClient.ListSourceConnectors(ctx, &connectorPB.ListSourceConnectorsRequest{
			PageToken: nextPageToken,
		})

		if err != nil {
			return err
		}

		nextPageToken = &resp.NextPageToken
		totalSize -= repository.DefaultPageSize
		connectors = append(connectors, resp.SourceConnectors...)
	}

	for _, connector := range connectors {
		resourceName := util.ConvertConnectorToResourceName(connector.Name)
		workflowId, _ := s.GetResourceWorkflowId(ctx, resourceName)
		// check if there is an ongoing workflow
		if workflowId != nil {
			opInfo, err := s.getOperationInfo(*workflowId, util.RESOURCE_TYPE_SOURCE_CONNECTOR)
			if err != nil {
				return err
			}
			if err := s.updateRunningConnector(ctx, resourceName, *opInfo); err != nil {
				return err
			}
			// if not trigger connector check workflow
		} else {
			resp, err := s.connectorPrivateClient.CheckSourceConnector(ctx, &connectorPB.CheckSourceConnectorRequest{
				Name: connector.Name,
			})
			if err != nil {
				return err
			}
			// non grpc/http connector, save workflowid
			if err := s.updateStaleConnector(ctx, resourceName, resp.WorkflowId); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *service) ProbeDestinationConnectors(ctx context.Context, cancel context.CancelFunc) error {
	defer cancel()

	resp, err := s.connectorPublicClient.ListDestinationConnectors(ctx, &connectorPB.ListDestinationConnectorsRequest{})

	if err != nil {
		return err
	}

	connectors := resp.DestinationConnectors
	nextPageToken := &resp.NextPageToken
	totalSize := resp.TotalSize

	for totalSize > repository.DefaultPageSize {
		resp, err := s.connectorPublicClient.ListDestinationConnectors(ctx, &connectorPB.ListDestinationConnectorsRequest{
			PageToken: nextPageToken,
		})

		if err != nil {
			return err
		}

		nextPageToken = &resp.NextPageToken
		totalSize -= repository.DefaultPageSize
		connectors = append(connectors, resp.DestinationConnectors...)
	}

	for _, connector := range connectors {
		resourceName := util.ConvertConnectorToResourceName(connector.Name)
		workflowId, _ := s.GetResourceWorkflowId(ctx, resourceName)
		// check if there is an ongoing workflow
		if workflowId != nil {
			opInfo, err := s.getOperationInfo(*workflowId, util.RESOURCE_TYPE_DESTINATION_CONNECTOR)
			if err != nil {
				return err
			}
			if err := s.updateRunningConnector(ctx, resourceName, *opInfo); err != nil {
				return err
			}
			// if not trigger connector check workflow
		} else {
			resp, err := s.connectorPrivateClient.CheckDestinationConnector(ctx, &connectorPB.CheckDestinationConnectorRequest{
				Name: connector.Name,
			})
			if err != nil {
				return err
			}
			// non grpc/http connector, save workflowid
			if err := s.updateStaleConnector(ctx, resourceName, resp.WorkflowId); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *service) updateRunningConnector(ctx context.Context, resourceName string, opInfo longrunningpb.Operation) error {
	logger, _ := logger.GetZapLogger()

	// if workflow done get result, otherwise remains same state
	if opInfo.Done {
		stateInt, err := strconv.ParseInt(string(opInfo.GetResponse().Value[:]), 10, 32)
		if err != nil {
			return err
		}
		if err := s.UpdateResourceState(ctx, &controllerPB.Resource{
			Name: resourceName,
			State: &controllerPB.Resource_ConnectorState{
				ConnectorState: connectorPB.Connector_State(stateInt),
			},
		}); err != nil {
			return err
		}
		if err := s.DeleteResourceWorkflowId(ctx, resourceName); err != nil {
			return err
		}
	}

	logResp, _ := s.GetResourceState(ctx, resourceName)
	logger.Info(fmt.Sprintf("[Controller] Got %v", logResp))

	return nil
}

func (s *service) updateStaleConnector(ctx context.Context, resourceName string, workflowId string) error {
	logger, _ := logger.GetZapLogger()

	if workflowId != "" {
		if err := s.UpdateResourceWorkflowId(ctx, resourceName, workflowId); err != nil {
			return err
		}
	} else {
		if err := s.UpdateResourceState(ctx, &controllerPB.Resource{
			Name: resourceName,
			State: &controllerPB.Resource_ConnectorState{
				ConnectorState: connectorPB.Connector_STATE_CONNECTED,
			},
		}); err != nil {
			return err
		}
	}

	logResp, _ := s.GetResourceState(ctx, resourceName)
	logger.Info(fmt.Sprintf("[Controller] Got %v", logResp))

	return nil
}
