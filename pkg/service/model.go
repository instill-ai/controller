package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/instill-ai/controller/internal/logger"
	"github.com/instill-ai/controller/internal/util"

	controllerPB "github.com/instill-ai/protogen-go/vdp/controller/v1alpha"
	modelPB "github.com/instill-ai/protogen-go/vdp/model/v1alpha"
)

func (s *service) ProbeModels(ctx context.Context, cancel context.CancelFunc) error {
	defer cancel()

	logger, _ := logger.GetZapLogger()

	var wg sync.WaitGroup

	resp, err := s.modelPrivateClient.ListModelsAdmin(ctx, &modelPB.ListModelsAdminRequest{})

	if err != nil {
		return err
	}

	models := resp.Models
	nextPageToken := &resp.NextPageToken
	totalSize := resp.TotalSize

	for totalSize > util.DefaultPageSize {
		resp, err := s.modelPrivateClient.ListModelsAdmin(ctx, &modelPB.ListModelsAdminRequest{
			PageToken: nextPageToken,
		})

		if err != nil {
			return err
		}

		nextPageToken = &resp.NextPageToken
		totalSize -= util.DefaultPageSize
		models = append(models, resp.Models...)
	}

	wg.Add(len(models))

	for _, model := range models {

		go func(model *modelPB.Model) {
			defer wg.Done()

			resourceName := util.ConvertRequestToResourceName(model.Name)

			workflowId, _ := s.GetResourceWorkflowId(ctx, resourceName)

			if workflowId != nil {
				opInfo, err := s.getOperationInfo(*workflowId, util.RESOURCE_TYPE_MODEL)
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
			} else {
				if resp, err := s.modelPrivateClient.CheckModel(ctx, &modelPB.CheckModelRequest{
					Name: model.Name,
				}); err == nil {
					if err = s.UpdateResourceState(ctx, &controllerPB.Resource{
						Name: resourceName,
						State: &controllerPB.Resource_ModelState{
							ModelState: resp.State,
						},
					}); err != nil {
						logger.Error(err.Error())
						return
					}
				} else {
					logger.Error(err.Error())
					return
				}
			}

			logResp, _ := s.GetResourceState(ctx, resourceName)
			logger.Info(fmt.Sprintf("[Controller] Got %v", logResp))
		}(model)

	}

	wg.Wait()

	return nil
}
