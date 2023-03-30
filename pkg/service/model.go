package service

import (
	"context"
	"fmt"

	"github.com/instill-ai/controller/internal/logger"
	"github.com/instill-ai/controller/internal/util"
	"github.com/instill-ai/model-backend/pkg/repository"

	controllerPB "github.com/instill-ai/protogen-go/vdp/controller/v1alpha"
	modelPB "github.com/instill-ai/protogen-go/vdp/model/v1alpha"
)

func (s *service) ProbeModels(ctx context.Context, cancel context.CancelFunc) error {
	defer cancel()

	logger, _ := logger.GetZapLogger()

	resp, err := s.modelPublicClient.ListModels(ctx, &modelPB.ListModelsRequest{})

	if err != nil {
		return err
	}

	models := resp.Models
	nextPageToken := &resp.NextPageToken
	totalSize := resp.TotalSize

	for totalSize > repository.DefaultPageSize {
		resp, err := s.modelPublicClient.ListModels(ctx, &modelPB.ListModelsRequest{
			PageToken: nextPageToken,
		})

		if err != nil {
			return err
		}

		nextPageToken = &resp.NextPageToken
		totalSize -= repository.DefaultPageSize
		models = append(models, resp.Models...)
	}

	modelInstances := []*modelPB.ModelInstance{}
	for _, model := range models {
		view := modelPB.View_VIEW_FULL
		resp, err := s.modelPublicClient.ListModelInstances(ctx, &modelPB.ListModelInstancesRequest{
			Parent: model.Name,
			View:   &view,
		})
		if err != nil {
			return err
		}

		nextPageToken := &resp.NextPageToken
		totalSize := resp.TotalSize
		modelInstances = append(modelInstances, resp.Instances...)

		for totalSize > repository.DefaultPageSize {
			resp, err := s.modelPublicClient.ListModelInstances(ctx, &modelPB.ListModelInstancesRequest{
				Parent:    model.Name,
				PageToken: nextPageToken,
				View:      &view,
			})

			if err != nil {
				return err
			}

			nextPageToken = &resp.NextPageToken
			totalSize -= repository.DefaultPageSize
			modelInstances = append(modelInstances, resp.Instances...)
		}
	}

	for _, modelInstance := range modelInstances {
		resp, err := s.modelPrivateClient.CheckModelInstance(ctx, &modelPB.CheckModelInstanceRequest{
			Name: modelInstance.Name,
		})

		if err != nil {
			return err
		}

		err = s.UpdateResourceState(ctx, &controllerPB.Resource{
			Name: util.ConvertModelToResourceName(modelInstance.Name),
			State: &controllerPB.Resource_ModelInstanceState{
				ModelInstanceState: resp.State,
			},
		})

		if err != nil {
			return err
		}

		logResp, _ := s.GetResourceState(ctx, util.ConvertModelToResourceName(modelInstance.Name))
		logger.Info(fmt.Sprintf("[Controller] Got %v", logResp))
	}
	return nil
}
