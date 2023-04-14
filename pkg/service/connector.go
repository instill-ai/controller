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
			} else {
				// if user desires connected
				resp, err := s.connectorPrivateClient.CheckSourceConnector(ctx, &connectorPB.CheckSourceConnectorRequest{
					Name: connector.Name,
				})
				if err != nil {
					logger.Error(err.Error())
					return
				}
				if err := s.UpdateResourceState(ctx, &controllerPB.Resource{
					Name: resourceName,
					State: &controllerPB.Resource_ConnectorState{
						ConnectorState: resp.State,
					},
				}); err != nil {
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

	resp, err := s.connectorPrivateClient.ListDestinationConnectorsAdmin(ctx, &connectorPB.ListDestinationConnectorsAdminRequest{})

	if err != nil {
		return err
	}

	connectors := resp.DestinationConnectors
	nextPageToken := &resp.NextPageToken
	totalSize := resp.TotalSize

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
			} else {
				// if user desires connected
				resp, err := s.connectorPrivateClient.CheckDestinationConnector(ctx, &connectorPB.CheckDestinationConnectorRequest{
					Name: connector.Name,
				})
				if err != nil {
					logger.Error(err.Error())
					return
				}
				if err := s.UpdateResourceState(ctx, &controllerPB.Resource{
					Name: resourceName,
					State: &controllerPB.Resource_ConnectorState{
						ConnectorState: resp.State,
					},
				}); err != nil {
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
