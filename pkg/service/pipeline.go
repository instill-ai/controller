package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/instill-ai/controller/internal/util"
	"github.com/instill-ai/controller/pkg/logger"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
	controllerPB "github.com/instill-ai/protogen-go/vdp/controller/v1alpha"
	modelPB "github.com/instill-ai/protogen-go/vdp/model/v1alpha"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
)

func (s *service) ProbePipelines(ctx context.Context, cancel context.CancelFunc) error {
	defer cancel()

	logger, _ := logger.GetZapLogger(ctx)

	var wg sync.WaitGroup

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

	resourceType := "pipelines"

	wg.Add(len(pipelines))

	for _, pipeline := range pipelines {

		go func(pipeline *pipelinePB.Pipeline) {
			defer wg.Done()

			resourcePermalink := util.ConvertUIDToResourcePermalink(pipeline.Uid, resourceType)

			pipelineResource := controllerPB.Resource{
				ResourcePermalink: resourcePermalink,
				State: &controllerPB.Resource_PipelineState{
					PipelineState: pipelinePB.Pipeline_STATE_INACTIVE,
				},
			}

			// user desires inactive
			if pipeline.State == pipelinePB.Pipeline_STATE_INACTIVE {
				if err := s.UpdateResourceState(ctx, &pipelineResource); err != nil {
					logger.Error(err.Error())
					return
				} else {
					return
				}
			}

			// user desires active, now check each component's state
			pipelineResource.State = &controllerPB.Resource_PipelineState{PipelineState: pipelinePB.Pipeline_STATE_ERROR}

			var resources []*controllerPB.Resource

			for _, component := range pipeline.Recipe.Components {

				if i := strings.Index(component.ResourceName, "/"); i >= 0 {
					switch component.ResourceName[:i] {
					case "source-connectors":
						sourceConnectorResource, err := s.GetResourceState(ctx, util.ConvertUIDToResourcePermalink(strings.Split(component.ResourceName, "/")[1], "source-connectors"))
						if err != nil {
							resErr := s.UpdateResourceState(ctx, &pipelineResource)
							if resErr != nil {
								logger.Error(fmt.Sprintf("UpdateResourceState failed for %s", component.ResourceName))
							}
							logger.Error(fmt.Sprintf("no record found for %s in etcd", component.ResourceName))
							return
						}
						resources = append(resources, sourceConnectorResource)
					case "destination-connectors":
						destinationConnectorResource, err := s.GetResourceState(ctx, util.ConvertUIDToResourcePermalink(strings.Split(component.ResourceName, "/")[1], "destination-connectors"))
						if err != nil {
							resErr := s.UpdateResourceState(ctx, &pipelineResource)
							if resErr != nil {
								logger.Error(fmt.Sprintf("UpdateResourceState failed for %s", component.ResourceName))
							}
							logger.Error(fmt.Sprintf("no record found for %s in etcd", component.ResourceName))
							return
						}
						resources = append(resources, destinationConnectorResource)
					case "models":
						modelResource, err := s.GetResourceState(ctx, util.ConvertUIDToResourcePermalink(strings.Split(component.ResourceName, "/")[1], "models"))
						if err != nil {
							resErr := s.UpdateResourceState(ctx, &pipelineResource)
							if resErr != nil {
								logger.Error(fmt.Sprintf("UpdateResourceState failed for %s", component.ResourceName))
							}
							logger.Error(fmt.Sprintf("no record found for %s in etcd", component.ResourceName))
							return
						}

						resources = append(resources, modelResource)
					}
				}

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
				resErr := s.UpdateResourceState(ctx, &pipelineResource)
				if resErr != nil {
					logger.Error(fmt.Sprintf("UpdateResourceState failed for %s", pipeline.Name))
				}
				return
			}

			pipelineResource.State = &controllerPB.Resource_PipelineState{
				PipelineState: pipelinePB.Pipeline_STATE_ACTIVE,
			}
			resErr := s.UpdateResourceState(ctx, &pipelineResource)
			if resErr != nil {
				logger.Error(fmt.Sprintf("UpdateResourceState failed for %s", pipeline.Name))
			}

			logResp, _ := s.GetResourceState(ctx, resourcePermalink)
			logger.Info(fmt.Sprintf("[Controller] Got %v", logResp))
		}(pipeline)
	}

	wg.Wait()

	return nil
}
