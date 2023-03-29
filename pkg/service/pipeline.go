package service

import (
	"context"
	"fmt"
	"time"

	"github.com/instill-ai/controller/config"
	"github.com/instill-ai/controller/internal/logger"
	"github.com/instill-ai/controller/internal/util"
	"github.com/instill-ai/model-backend/pkg/repository"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
	controllerPB "github.com/instill-ai/protogen-go/vdp/controller/v1alpha"
	modelPB "github.com/instill-ai/protogen-go/vdp/model/v1alpha"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
)

func (s *service) ProbePipelines() error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Config.Server.Timeout*time.Second)
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

	for totalSize > repository.DefaultPageSize {
		resp, err := s.pipelinePrivateClient.ListPipelinesAdmin(ctx, &pipelinePB.ListPipelinesAdminRequest{
			PageToken: nextPageToken,
			View: pipelinePB.View_VIEW_FULL.Enum(),
		})

		if err != nil {
			return err
		}

		nextPageToken = &resp.NextPageToken
		totalSize -= repository.DefaultPageSize
		pipelines = append(pipelines, resp.Pipelines...)
	}

	for _, pipeline := range pipelines {
		resourceName := util.ConvertPipelineToResourceName(pipeline.Name)

		pipelineResource := controllerPB.Resource{
			Name: resourceName,
			State: &controllerPB.Resource_PipelineState{
				PipelineState: pipelinePB.Pipeline_STATE_ERROR,
			},
		}

		var resources []*controllerPB.Resource

		sourceConnectorResource, err := s.GetResourceState(util.ConvertConnectorToResourceName(pipeline.Recipe.Source))
		if err != nil {
			s.UpdateResourceState(&pipelineResource)
			logger.Error("no record find for source connector")
			return err
		}
		resources = append(resources, sourceConnectorResource)

		destinationConnectorResource, err := s.GetResourceState(util.ConvertConnectorToResourceName(pipeline.Recipe.Destination))
		if err != nil {
			s.UpdateResourceState(&pipelineResource)
			logger.Error("no record find for destination connector")
			return err
		}
		resources = append(resources, destinationConnectorResource)

		modelInstanceNames := pipeline.Recipe.ModelInstances
		for _, modelInstanceName := range modelInstanceNames {
			modelInstanceResource, err := s.GetResourceState(util.ConvertModelToResourceName(modelInstanceName))
			if err != nil {
				s.UpdateResourceState(&pipelineResource)
				logger.Error(fmt.Sprintf("no record find for model instance %v", modelInstanceName))
				return err
			}

			resources = append(resources, modelInstanceResource)
		}

		for _, r := range resources {
			switch v := r.State.(type){
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
			case *controllerPB.Resource_ModelInstanceState:
				switch v.ModelInstanceState {
				case modelPB.ModelInstance_STATE_OFFLINE:
					pipelineResource.State = &controllerPB.Resource_PipelineState{
						PipelineState: pipelinePB.Pipeline_STATE_INACTIVE,
					}
				case modelPB.ModelInstance_STATE_UNSPECIFIED:
					pipelineResource.State = &controllerPB.Resource_PipelineState{
						PipelineState: pipelinePB.Pipeline_STATE_UNSPECIFIED,
					}
				case modelPB.ModelInstance_STATE_ERROR:
					pipelineResource.State = &controllerPB.Resource_PipelineState{
						PipelineState: pipelinePB.Pipeline_STATE_ERROR,
					}
				default:
					continue
				}
			}
			s.UpdateResourceState(&pipelineResource)
			return nil
		}

		pipelineResource.State = &controllerPB.Resource_PipelineState{
			PipelineState: pipelinePB.Pipeline_STATE_ACTIVE,
		}
		s.UpdateResourceState(&pipelineResource)

		logResp, _ := s.GetResourceState(resourceName)
		logger.Info(fmt.Sprintf("[Controller] Got %v", logResp))
	}
	return nil
}
