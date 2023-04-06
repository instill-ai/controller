package service

import (
	"context"
	"fmt"

	"github.com/instill-ai/controller/internal/logger"
	"github.com/instill-ai/controller/internal/util"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
	controllerPB "github.com/instill-ai/protogen-go/vdp/controller/v1alpha"
	modelPB "github.com/instill-ai/protogen-go/vdp/model/v1alpha"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
)

func (s *service) ProbePipelines(ctx context.Context, cancel context.CancelFunc) error {
	defer cancel()

	logger, _ := logger.GetZapLogger()

	resp, err := s.pipelinePrivateClient.ListPipelinesAdmin(ctx, &pipelinePB.ListPipelinesAdminRequest{
		View: pipelinePB.View_VIEW_FULL.Enum(),
	})

	if err != nil {
		return err
	}

	pipelines := resp.Pipelines
	nextPageToken := &resp.NextPageToken
	totalSize := resp.TotalSize

	for totalSize > util.DefaultPageSize {
		resp, err := s.pipelinePrivateClient.ListPipelinesAdmin(ctx, &pipelinePB.ListPipelinesAdminRequest{
			PageToken: nextPageToken,
			View:      pipelinePB.View_VIEW_FULL.Enum(),
		})

		if err != nil {
			return err
		}

		nextPageToken = &resp.NextPageToken
		totalSize -= util.DefaultPageSize
		pipelines = append(pipelines, resp.Pipelines...)
	}

	for _, pipeline := range pipelines {
		resourceName := util.ConvertRequestToResourceName(pipeline.Name)

		pipelineResource := controllerPB.Resource{
			Name: resourceName,
			State: &controllerPB.Resource_PipelineState{
				PipelineState: pipelinePB.Pipeline_STATE_INACTIVE,
			},
		}

		// user desires inactive
		if pipeline.State == pipelinePB.Pipeline_STATE_INACTIVE {
			if err := s.UpdateResourceState(ctx, &pipelineResource); err != nil {
				return err
			} else {
				return nil
			}
		}

		// user desires active, now check each component's state
		pipelineResource.State = &controllerPB.Resource_PipelineState{PipelineState: pipelinePB.Pipeline_STATE_ERROR}

		var resources []*controllerPB.Resource

		sourceConnectorResource, err := s.GetResourceState(ctx, util.ConvertRequestToResourceName(pipeline.Recipe.Source))
		if err != nil {
			s.UpdateResourceState(ctx, &pipelineResource)
			logger.Error("no record find for source connector")
			return err
		}
		resources = append(resources, sourceConnectorResource)

		destinationConnectorResource, err := s.GetResourceState(ctx, util.ConvertRequestToResourceName(pipeline.Recipe.Destination))
		if err != nil {
			s.UpdateResourceState(ctx, &pipelineResource)
			logger.Error("no record find for destination connector")
			return err
		}
		resources = append(resources, destinationConnectorResource)

		modelNames := pipeline.Recipe.Models
		for _, modelName := range modelNames {
			modelResource, err := s.GetResourceState(ctx, util.ConvertRequestToResourceName(modelName))
			if err != nil {
				s.UpdateResourceState(ctx, &pipelineResource)
				logger.Error(fmt.Sprintf("no record find for model  %v", modelName))
				return err
			}

			resources = append(resources, modelResource)
		}

		for _, r := range resources {
			switch v := r.State.(type) {
			case *controllerPB.Resource_ConnectorState:
				switch v.ConnectorState {
				case connectorPB.Connector_STATE_DISCONNECTED:
					pipelineResource.State = &controllerPB.Resource_PipelineState{
						PipelineState: pipelinePB.Pipeline_STATE_INACTIVE,
					}
				case connectorPB.Connector_STATE_UNSPECIFIED:
					pipelineResource.State = &controllerPB.Resource_PipelineState{
						PipelineState: pipelinePB.Pipeline_STATE_UNSPECIFIED,
					}
				case connectorPB.Connector_STATE_ERROR:
					pipelineResource.State = &controllerPB.Resource_PipelineState{
						PipelineState: pipelinePB.Pipeline_STATE_ERROR,
					}
				default:
					continue
				}
			case *controllerPB.Resource_ModelState:
				switch v.ModelState {
				case modelPB.Model_STATE_OFFLINE:
					pipelineResource.State = &controllerPB.Resource_PipelineState{
						PipelineState: pipelinePB.Pipeline_STATE_INACTIVE,
					}
				case modelPB.Model_STATE_UNSPECIFIED:
					pipelineResource.State = &controllerPB.Resource_PipelineState{
						PipelineState: pipelinePB.Pipeline_STATE_UNSPECIFIED,
					}
				case modelPB.Model_STATE_ERROR:
					pipelineResource.State = &controllerPB.Resource_PipelineState{
						PipelineState: pipelinePB.Pipeline_STATE_ERROR,
					}
				default:
					continue
				}
			}
			s.UpdateResourceState(ctx, &pipelineResource)
			return nil
		}

		pipelineResource.State = &controllerPB.Resource_PipelineState{
			PipelineState: pipelinePB.Pipeline_STATE_ACTIVE,
		}
		s.UpdateResourceState(ctx, &pipelineResource)

		logResp, _ := s.GetResourceState(ctx, resourceName)
		logger.Info(fmt.Sprintf("[Controller] Got %v", logResp))
	}
	return nil
}
