package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/instill-ai/controller/internal/logger"
	"github.com/instill-ai/controller/internal/util"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
	controllerPB "github.com/instill-ai/protogen-go/vdp/controller/v1alpha"
)

func (s *service) ProbeSourceConnectors(ctx context.Context, cancel context.CancelFunc) error {
	defer cancel()

	logger, _ := logger.GetZapLogger()

	var wg sync.WaitGroup

	resp, err := s.connectorPublicClient.ListSourceConnectors(ctx, &connectorPB.ListSourceConnectorsRequest{})

	if err != nil {
		return err
	}

	connectors := resp.SourceConnectors
	nextPageToken := &resp.NextPageToken
	totalSize := resp.TotalSize

	wg.Add(int(totalSize))

	for totalSize > util.DefaultPageSize {
		resp, err := s.connectorPublicClient.ListSourceConnectors(ctx, &connectorPB.ListSourceConnectorsRequest{
			PageToken: nextPageToken,
		})

		if err != nil {
			return err
		}

		nextPageToken = &resp.NextPageToken
		totalSize -= util.DefaultPageSize
		connectors = append(connectors, resp.SourceConnectors...)
	}

	for _, connector := range connectors {

		go func(connector *connectorPB.SourceConnector) {
			defer wg.Done()

			resourceName := util.ConvertRequestToResourceName(connector.Name)

			// if user desires disconnected
			if connector.Connector.State == connectorPB.Connector_STATE_DISCONNECTED {
				if err := s.UpdateResourceState(ctx, &controllerPB.Resource{
					Name: resourceName,
					State: &controllerPB.Resource_ConnectorState{
						ConnectorState: connectorPB.Connector_STATE_DISCONNECTED,
					},
				}); err != nil {
					logger.Error(err.Error())
					return
				}
			}
			// if user desires connected
			workflowId, _ := s.GetResourceWorkflowId(ctx, resourceName)
			// check if there is an ongoing workflow
			if workflowId != nil {
				opInfo, err := s.getOperationInfo(*workflowId, util.RESOURCE_TYPE_SOURCE_CONNECTOR)
				if err != nil {
					logger.Error(err.Error())
					return
				}
				if opInfo.Done {
					if err := s.DeleteResourceWorkflowId(ctx, resourceName); err != nil {
						logger.Error(err.Error())
						return
					}
				}
				// if not trigger connector check workflow
			} else {
				resp, err := s.connectorPrivateClient.CheckSourceConnector(ctx, &connectorPB.CheckSourceConnectorRequest{
					Name: connector.Name,
				})
				if err != nil {
					logger.Error(err.Error())
					return
				}
				// non grpc/http connector, save workflowid
				if err := s.updateStaleConnector(ctx, resourceName, resp.WorkflowId); err != nil {
					logger.Error(err.Error())
					return
				}
			}
			logResp, _ := s.GetResourceState(ctx, resourceName)
			logger.Info(fmt.Sprintf("[Controller] Got %v", logResp))
		}(connector)
	}

	wg.Wait()

	return nil
}

func (s *service) ProbeDestinationConnectors(ctx context.Context, cancel context.CancelFunc) error {
	defer cancel()

	logger, _ := logger.GetZapLogger()

	var wg sync.WaitGroup

	resp, err := s.connectorPublicClient.ListDestinationConnectors(ctx, &connectorPB.ListDestinationConnectorsRequest{})

	if err != nil {
		return err
	}

	connectors := resp.DestinationConnectors
	nextPageToken := &resp.NextPageToken
	totalSize := resp.TotalSize

	for totalSize > util.DefaultPageSize {
		resp, err := s.connectorPublicClient.ListDestinationConnectors(ctx, &connectorPB.ListDestinationConnectorsRequest{
			PageToken: nextPageToken,
		})

		if err != nil {
			return err
		}

		nextPageToken = &resp.NextPageToken
		totalSize -= util.DefaultPageSize
		connectors = append(connectors, resp.DestinationConnectors...)
	}

	wg.Add(len(connectors))

	for _, connector := range connectors {

		go func(connector *connectorPB.DestinationConnector) {
			defer wg.Done()

			resourceName := util.ConvertRequestToResourceName(connector.Name)

			// if user desires disconnected
			if connector.Connector.State == connectorPB.Connector_STATE_DISCONNECTED {
				if err := s.UpdateResourceState(ctx, &controllerPB.Resource{
					Name: resourceName,
					State: &controllerPB.Resource_ConnectorState{
						ConnectorState: connectorPB.Connector_STATE_DISCONNECTED,
					},
				}); err != nil {
					logger.Error(err.Error())
					return
				}
			}
			// if user desires connected
			workflowId, _ := s.GetResourceWorkflowId(ctx, resourceName)
			// check if there is an ongoing workflow
			if workflowId != nil {
				opInfo, err := s.getOperationInfo(*workflowId, util.RESOURCE_TYPE_DESTINATION_CONNECTOR)
				if err != nil {
					logger.Error(err.Error())
					return
				}
				if opInfo.Done {
					if err := s.DeleteResourceWorkflowId(ctx, resourceName); err != nil {
						logger.Error(err.Error())
						return
					}
				}
				// if not trigger connector check workflow
			} else {
				resp, err := s.connectorPrivateClient.CheckDestinationConnector(ctx, &connectorPB.CheckDestinationConnectorRequest{
					Name: connector.Name,
				})
				if err != nil {
					logger.Error(err.Error())
					return
				}
				if err := s.updateStaleConnector(ctx, resourceName, resp.WorkflowId); err != nil {
					logger.Error(err.Error())
					return
				}
			}
			logResp, _ := s.GetResourceState(ctx, resourceName)
			logger.Info(fmt.Sprintf("[Controller] Got %v", logResp))
		}(connector)
	}

	wg.Wait()

	return nil
}

func (s *service) updateStaleConnector(ctx context.Context, resourceName string, workflowId string) error {
	// non grpc/http connector, save workflowid
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

	return nil
}
