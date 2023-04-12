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

	wg.Add(len(models))

	for _, model := range models {

		go func(model *modelPB.Model) {
			defer wg.Done()

			if resp, err := s.modelPrivateClient.CheckModel(ctx, &modelPB.CheckModelRequest{
				Name: model.Name,
			}); err == nil {
				if err = s.UpdateResourceState(ctx, &controllerPB.Resource{
					Name: util.ConvertRequestToResourceName(model.Name),
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

			logResp, _ := s.GetResourceState(ctx, util.ConvertRequestToResourceName(model.Name))
			logger.Info(fmt.Sprintf("[Controller] Got %v", logResp))
		}(model)

	}

	wg.Wait()

	return nil
}
