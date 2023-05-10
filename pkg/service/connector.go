package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/instill-ai/controller/internal/logger"
	"github.com/instill-ai/controller/internal/util"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
	controllerPB "github.com/instill-ai/protogen-go/vdp/controller/v1alpha"
)

func (s *service) ProbeSourceConnectors(ctx context.Context, cancel context.CancelFunc) error {
	defer cancel()

	logger, _ := logger.GetZapLogger()

	var wg sync.WaitGroup

	resp, err := s.connectorPrivateClient.ListSourceConnectorsAdmin(ctx, &connectorPB.ListSourceConnectorsAdminRequest{})

	if err != nil {
		return err
	}

	connectors := resp.SourceConnectors
	nextPageToken := &resp.NextPageToken
	totalSize := resp.TotalSize

	wg.Add(int(totalSize))

	for totalSize > util.DefaultPageSize {
		resp, err := s.connectorPrivateClient.ListSourceConnectorsAdmin(ctx, &connectorPB.ListSourceConnectorsAdminRequest{
			PageToken: nextPageToken,
		})

		if err != nil {
			return err
		}

		nextPageToken = &resp.NextPageToken
		totalSize -= util.DefaultPageSize
		connectors = append(connectors, resp.SourceConnectors...)
	}

	connectorType := "source-connectors"

	for _, connector := range connectors {

		go func(connector *connectorPB.SourceConnector) {
			defer wg.Done()

			resourcePermalink := util.ConvertUIDToResourcePermalink(connector.Uid, connectorType)

			// if user desires disconnected
			if connector.Connector.State == connectorPB.Connector_STATE_DISCONNECTED {
				if err := s.UpdateResourceState(ctx, &controllerPB.Resource{
					ResourcePermalink: resourcePermalink,
					State: &controllerPB.Resource_ConnectorState{
						ConnectorState: connectorPB.Connector_STATE_DISCONNECTED,
					},
				}); err != nil {
					logger.Error(err.Error())
					return
				}
			}
			// if user desires connected
			workflowId, _ := s.GetResourceWorkflowId(ctx, resourcePermalink)
			// check if there is an ongoing workflow
			if workflowId != nil {
				opInfo, err := s.getOperationInfo(*workflowId, util.RESOURCE_TYPE_SOURCE_CONNECTOR)
				if err != nil {
					logger.Error(err.Error())
					return
				}
				if opInfo.Done {
					if err := s.DeleteResourceWorkflowId(ctx, resourcePermalink); err != nil {
						logger.Error(err.Error())
						return
					}
				}
				// if not trigger connector check workflow
			} else {
				resp, err := s.connectorPrivateClient.CheckSourceConnector(ctx, &connectorPB.CheckSourceConnectorRequest{
					SourceConnectorPermalink: fmt.Sprintf("%s/%s", connectorType, connector.Uid),
				})
				if err != nil {
					logger.Error(err.Error())
					return
				}
				// non grpc/http connector, save workflowid
				if err := s.updateStaleConnector(ctx, resourcePermalink, resp.WorkflowId); err != nil {
					logger.Error(err.Error())
					return
				}
			}
			// Temporary block for repopulate resources
			workflowId, _ = s.GetResourceWorkflowId(ctx, resourcePermalink)
			if workflowId != nil {
				logger.Info(fmt.Sprintf("[controller] repopulate %s...", resourcePermalink))
				for {
					opInfo, err := s.getOperationInfo(*workflowId, util.RESOURCE_TYPE_SOURCE_CONNECTOR)
					if err != nil {
						logger.Error(err.Error())
					}
					if opInfo.Done{
						if err := s.DeleteResourceWorkflowId(ctx, resourcePermalink); err != nil {
							logger.Error(err.Error())
							return
						}
						break
					}
					time.Sleep(time.Second)
				}
			}

			logResp, _ := s.GetResourceState(ctx, resourcePermalink)
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

	resp, err := s.connectorPrivateClient.ListDestinationConnectorsAdmin(ctx, &connectorPB.ListDestinationConnectorsAdminRequest{})

	if err != nil {
		return err
	}

	connectors := resp.DestinationConnectors
	nextPageToken := &resp.NextPageToken
	totalSize := resp.TotalSize

	wg.Add(int(totalSize))

	for totalSize > util.DefaultPageSize {
		resp, err := s.connectorPrivateClient.ListDestinationConnectorsAdmin(ctx, &connectorPB.ListDestinationConnectorsAdminRequest{
			PageToken: nextPageToken,
		})

		if err != nil {
			return err
		}

		nextPageToken = &resp.NextPageToken
		totalSize -= util.DefaultPageSize
		connectors = append(connectors, resp.DestinationConnectors...)
	}

	connectorType := "destination-connectors"

	for _, connector := range connectors {

		go func(connector *connectorPB.DestinationConnector) {
			defer wg.Done()

			resourcePermalink := util.ConvertUIDToResourcePermalink(connector.Uid, connectorType)

			// if user desires disconnected
			if connector.Connector.State == connectorPB.Connector_STATE_DISCONNECTED {
				if err := s.UpdateResourceState(ctx, &controllerPB.Resource{
					ResourcePermalink: resourcePermalink,
					State: &controllerPB.Resource_ConnectorState{
						ConnectorState: connectorPB.Connector_STATE_DISCONNECTED,
					},
				}); err != nil {
					logger.Error(err.Error())
					return
				}
			}
			// if user desires connected
			workflowId, _ := s.GetResourceWorkflowId(ctx, resourcePermalink)
			// check if there is an ongoing workflow
			if workflowId != nil {
				opInfo, err := s.getOperationInfo(*workflowId, util.RESOURCE_TYPE_DESTINATION_CONNECTOR)
				if err != nil {
					logger.Error(err.Error())
					return
				}
				if opInfo.Done {
					if err := s.DeleteResourceWorkflowId(ctx, resourcePermalink); err != nil {
						logger.Error(err.Error())
						return
					}
				}
				// if not trigger connector check workflow
			} else {
				resp, err := s.connectorPrivateClient.CheckDestinationConnector(ctx, &connectorPB.CheckDestinationConnectorRequest{
					DestinationConnectorPermalink: fmt.Sprintf("%s/%s", connectorType, connector.Uid),
				})
				if err != nil {
					logger.Error(err.Error())
					return
				}
				if err := s.updateStaleConnector(ctx, resourcePermalink, resp.WorkflowId); err != nil {
					logger.Error(err.Error())
					return
				}
			}
			// Temporary block for repopulate resources
			workflowId, _ = s.GetResourceWorkflowId(ctx, resourcePermalink)
			if workflowId != nil {
				logger.Info(fmt.Sprintf("[controller] repopulate %s...", resourcePermalink))
				for {
					opInfo, err := s.getOperationInfo(*workflowId, util.RESOURCE_TYPE_DESTINATION_CONNECTOR)
					if err != nil {
						logger.Error(err.Error())
					}
					if opInfo.Done{
						if err := s.DeleteResourceWorkflowId(ctx, resourcePermalink); err != nil {
							logger.Error(err.Error())
							return
						}
						break
					}
					time.Sleep(time.Second)
				}
			}
			logResp, _ := s.GetResourceState(ctx, resourcePermalink)
			logger.Info(fmt.Sprintf("[Controller] Got %v", logResp))
		}(connector)
	}

	wg.Wait()

	return nil
}

func (s *service) updateStaleConnector(ctx context.Context, resourcePermalink string, workflowId string) error {
	// non grpc/http connector, save workflowid
	if workflowId != "" {
		if err := s.UpdateResourceWorkflowId(ctx, resourcePermalink, workflowId); err != nil {
			return err
		}
	} else {
		if err := s.UpdateResourceState(ctx, &controllerPB.Resource{
			ResourcePermalink: resourcePermalink,
			State: &controllerPB.Resource_ConnectorState{
				ConnectorState: connectorPB.Connector_STATE_CONNECTED,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}
