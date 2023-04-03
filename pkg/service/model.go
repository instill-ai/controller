package service

import (
	"context"
	"fmt"

	"github.com/instill-ai/controller/internal/logger"
	"github.com/instill-ai/controller/internal/util"

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

	for totalSize > util.DefaultPageSize {
		resp, err := s.modelPublicClient.ListModels(ctx, &modelPB.ListModelsRequest{
			PageToken: nextPageToken,
		})

		if err != nil {
			return err
		}

		nextPageToken = &resp.NextPageToken
		totalSize -= util.DefaultPageSize
		models = append(models, resp.Models...)
	}

	for _, model := range models {
		resp, err := s.modelPrivateClient.CheckModel(ctx, &modelPB.CheckModelRequest{
			Name: model.Name,
		})

		if err != nil {
			return err
		}

		err = s.UpdateResourceState(ctx, &controllerPB.Resource{
			Name: util.ConvertRequestToResourceName(model.Name),
			State: &controllerPB.Resource_ModelState{
				ModelState: resp.State,
			},
		})

		if err != nil {
			return err
		}

		logResp, _ := s.GetResourceState(ctx, util.ConvertRequestToResourceName(model.Name))
		logger.Info(fmt.Sprintf("[Controller] Got %v", logResp))
	}
	return nil
}
