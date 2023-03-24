package repository

import (
	"time"

	"github.com/gofrs/uuid"
	"go.einride.tech/aip/filtering"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/instill-ai/x/paginate"
	modelDM "github.com/instill-ai/model-backend/pkg/datamodel"
	pipelineDM "github.com/instill-ai/pipeline-backend/pkg/datamodel"
	modelPB "github.com/instill-ai/protogen-go/vdp/model/v1alpha"
)

// DefaultPageSize is the default pagination page size when page size is not assigned
const DefaultPageSize = 10

// MaxPageSize is the maximum pagination page size if the assigned value is over this number
const MaxPageSize = 100

// Repository interface
type Repository interface {
	GetModelById(owner string, modelID string, view modelPB.View) (modelDM.Model, error)
	GetModelByUid(owner string, modelUID uuid.UUID, view modelPB.View) (modelDM.Model, error)
	GetModelInstance(modelUID uuid.UUID, instanceID string, view modelPB.View) (modelDM.ModelInstance, error)
	GetModelInstanceByUid(modelUID uuid.UUID, modelInstanceUid uuid.UUID, view modelPB.View) (modelDM.ModelInstance, error)
	GetModelInstances(modelUID uuid.UUID) ([]modelDM.ModelInstance, error)
	ListModelInstances(modelUID uuid.UUID, view modelPB.View, pageSize int, pageToken string) (instances []modelDM.ModelInstance, nextPageToken string, totalSize int64, err error)
	ListModels(owner string, view modelPB.View, pageSize int, pageToken string) (models []modelDM.Model, nextPageToken string, totalSize int64, err error)
	ListPipeline(owner string, pageSize int64, pageToken string, isBasicView bool, filter filtering.Filter) ([]pipelineDM.Pipeline, int64, string, error)
	GetPipelineByID(id string, owner string, isBasicView bool) (*pipelineDM.Pipeline, error)
	GetPipelineByUID(uid uuid.UUID, owner string, isBasicView bool) (*pipelineDM.Pipeline, error)
}

type repository struct {
	db *gorm.DB
}

var GetModelSelectedFields = []string{
	`CONCAT('models/', id) as name`,
	`"model"."uid"`,
	`"model"."id"`,
	`"model"."description"`,
	`"model"."model_definition_uid"`,
	`"model"."configuration"`,
	`"model"."visibility"`,
	`"model"."owner"`,
	`"model"."create_time"`,
	`"model"."update_time"`,
}

var GetModelSelectedFieldsWOConfiguration = []string{
	`CONCAT('models/', id) as name`,
	`"model"."uid"`,
	`"model"."id"`,
	`"model"."description"`,
	`"model"."model_definition_uid"`,
	`"model"."visibility"`,
	`"model"."owner"`,
	`"model"."create_time"`,
	`"model"."update_time"`,
}

// NewRepository initiates a repository instance
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) GetModelById(owner string, modelID string, view modelPB.View) (modelDM.Model, error) {
	var model modelDM.Model
	selectedFields := GetModelSelectedFields
	if view != modelPB.View_VIEW_FULL {
		selectedFields = GetModelSelectedFieldsWOConfiguration
	}
	if result := r.db.Model(&modelDM.Model{}).Select(selectedFields).Where(&modelDM.Model{Owner: owner, ID: modelID}).First(&model); result.Error != nil {
		return modelDM.Model{}, status.Errorf(codes.NotFound, "The model id %s you specified is not found in namespace %s", modelID, owner)
	}
	return model, nil
}

func (r *repository) GetModelByUid(owner string, modelUID uuid.UUID, view modelPB.View) (modelDM.Model, error) {
	var model modelDM.Model
	selectedFields := GetModelSelectedFields
	if view != modelPB.View_VIEW_FULL {
		selectedFields = GetModelSelectedFieldsWOConfiguration
	}
	if result := r.db.Model(&modelDM.Model{}).Select(selectedFields).Where(&modelDM.Model{Owner: owner, BaseDynamic: modelDM.BaseDynamic{UID: modelUID}}).First(&model); result.Error != nil {
		return modelDM.Model{}, status.Errorf(codes.NotFound, "The model uid %s you specified is not found in namespace %s", modelUID, owner)
	}
	return model, nil
}

func (r *repository) ListModels(owner string, view modelPB.View, pageSize int, pageToken string) (models []modelDM.Model, nextPageToken string, totalSize int64, err error) {
	if result := r.db.Model(&modelDM.Model{}).Where("owner = ?", owner).Count(&totalSize); result.Error != nil {
		return nil, "", 0, status.Errorf(codes.Internal, result.Error.Error())
	}

	queryBuilder := r.db.Model(&modelDM.Model{}).Order("create_time DESC, uid DESC").Where("owner = ?", owner)

	if pageSize == 0 {
		pageSize = DefaultPageSize
	} else if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	queryBuilder = queryBuilder.Limit(int(pageSize))

	if pageToken != "" {
		createTime, uid, err := paginate.DecodeToken(pageToken)
		if err != nil {
			return nil, "", 0, status.Errorf(codes.InvalidArgument, "Invalid page token: %s", err.Error())
		}
		queryBuilder = queryBuilder.Where("(create_time,uid) < (?::timestamp, ?)", createTime, uid)
	}

	if view != modelPB.View_VIEW_FULL {
		queryBuilder.Omit("configuration")
	}

	var createTime time.Time
	rows, err := queryBuilder.Rows()
	if err != nil {
		return nil, "", 0, status.Errorf(codes.Internal, err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var item modelDM.Model
		if err = r.db.ScanRows(rows, &item); err != nil {
			return nil, "", 0, status.Error(codes.Internal, err.Error())
		}
		createTime = item.CreateTime
		models = append(models, item)
	}

	if len(models) > 0 {
		lastUID := (models)[len(models)-1].UID
		lastItem := &modelDM.Model{}
		if result := r.db.Model(&modelDM.Model{}).
			Where("owner = ?", owner).
			Order("create_time ASC, uid ASC").
			Limit(1).Find(lastItem); result.Error != nil {
			return nil, "", 0, status.Errorf(codes.Internal, result.Error.Error())
		}
		if lastItem.UID.String() == lastUID.String() {
			nextPageToken = ""
		} else {
			nextPageToken = paginate.EncodeToken(createTime, lastUID.String())
		}
	}

	return models, nextPageToken, totalSize, nil
}

func (r *repository) GetModelInstance(modelUID uuid.UUID, instanceID string, view modelPB.View) (modelDM.ModelInstance, error) {
	var instanceDB modelDM.ModelInstance
	omit := ""
	if view != modelPB.View_VIEW_FULL {
		omit = "configuration"
	}
	if result := r.db.Model(&modelDM.ModelInstance{}).Omit(omit).Where(map[string]interface{}{"model_uid": modelUID, "id": instanceID}).First(&instanceDB); result.Error != nil {
		return modelDM.ModelInstance{}, status.Errorf(codes.NotFound, "The instance %v for model %v not found", instanceID, modelUID)
	}
	return instanceDB, nil
}

func (r *repository) GetModelInstanceByUid(modelUID uuid.UUID, modelInstanceUid uuid.UUID, view modelPB.View) (modelDM.ModelInstance, error) {
	var instanceDB modelDM.ModelInstance
	omit := ""
	if view != modelPB.View_VIEW_FULL {
		omit = "configuration"
	}
	if result := r.db.Model(&modelDM.ModelInstance{}).Omit(omit).Where(map[string]interface{}{"model_uid": modelUID, "uid": modelInstanceUid}).First(&instanceDB); result.Error != nil {
		return modelDM.ModelInstance{}, status.Errorf(codes.NotFound, "The instance uid %v for model uid %v not found", modelInstanceUid, modelUID)
	}
	return instanceDB, nil
}

func (r *repository) GetModelInstances(modelUID uuid.UUID) ([]modelDM.ModelInstance, error) {
	var instances []modelDM.ModelInstance
	if result := r.db.Model(&modelDM.ModelInstance{}).Where("model_uid", modelUID).Order("id asc").Find(&instances); result.Error != nil {
		return []modelDM.ModelInstance{}, status.Errorf(codes.NotFound, "The instance for model %v not found", modelUID)
	}
	return instances, nil
}

func (r *repository) ListModelInstances(modelUID uuid.UUID, view modelPB.View, pageSize int, pageToken string) (instances []modelDM.ModelInstance, nextPageToken string, totalSize int64, err error) {

	if result := r.db.Model(&modelDM.ModelInstance{}).Where("model_uid = ?", modelUID).Count(&totalSize); result.Error != nil {
		return nil, "", 0, status.Errorf(codes.Internal, result.Error.Error())
	}

	queryBuilder := r.db.Model(&modelDM.ModelInstance{}).Order("create_time DESC, uid DESC").Where("model_uid = ?", modelUID)

	if pageSize == 0 {
		pageSize = DefaultPageSize
	} else if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	queryBuilder = queryBuilder.Limit(int(pageSize))

	if pageToken != "" {
		createTime, uid, err := paginate.DecodeToken(pageToken)
		if err != nil {
			return nil, "", 0, status.Errorf(codes.InvalidArgument, "Invalid page token: %s", err.Error())
		}
		queryBuilder = queryBuilder.Where("(create_time,uid) < (?::timestamp, ?)", createTime, uid)
	}

	if view != modelPB.View_VIEW_FULL {
		queryBuilder.Omit("configuration")
	}

	var createTime time.Time
	rows, err := queryBuilder.Rows()
	if err != nil {
		return nil, "", 0, status.Errorf(codes.Internal, err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var item modelDM.ModelInstance
		if err = r.db.ScanRows(rows, &item); err != nil {
			return nil, "", 0, status.Error(codes.Internal, err.Error())
		}
		createTime = item.CreateTime
		instances = append(instances, item)
	}

	if len(instances) > 0 {
		lastUID := (instances)[len(instances)-1].UID
		lastItem := &modelDM.ModelInstance{}
		if result := r.db.Model(&modelDM.ModelInstance{}).
			Where("model_uid = ?", modelUID).
			Order("create_time ASC, uid ASC").
			Limit(1).Find(lastItem); result.Error != nil {
			return nil, "", 0, status.Errorf(codes.Internal, result.Error.Error())
		}
		if lastItem.UID.String() == lastUID.String() {
			nextPageToken = ""
		} else {
			nextPageToken = paginate.EncodeToken(createTime, lastUID.String())
		}
	}

	return instances, nextPageToken, totalSize, nil
}

func (r *repository) ListPipeline(owner string, pageSize int64, pageToken string, isBasicView bool, filter filtering.Filter) (pipelines []pipelineDM.Pipeline, totalSize int64, nextPageToken string, err error) {

	if result := r.db.Model(&pipelineDM.Pipeline{}).Where("owner = ?", owner).Count(&totalSize); result.Error != nil {
		return nil, 0, "", status.Errorf(codes.Internal, result.Error.Error())
	}

	queryBuilder := r.db.Model(&pipelineDM.Pipeline{}).Order("create_time DESC, uid DESC").Where("owner = ?", owner)

	if pageSize == 0 {
		pageSize = DefaultPageSize
	} else if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	queryBuilder = queryBuilder.Limit(int(pageSize))

	if pageToken != "" {
		createTime, uid, err := paginate.DecodeToken(pageToken)
		if err != nil {
			return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid page token: %s", err.Error())
		}
		queryBuilder = queryBuilder.Where("(create_time,uid) < (?::timestamp, ?)", createTime, uid)
	}

	if isBasicView {
		queryBuilder.Omit("pipeline.recipe")
	}

	if expr, err := r.transpileFilter(filter); err != nil {
		return nil, 0, "", status.Errorf(codes.Internal, err.Error())
	} else if expr != nil {
		queryBuilder.Clauses(expr)
	}

	var createTime time.Time
	rows, err := queryBuilder.Rows()
	if err != nil {
		return nil, 0, "", status.Errorf(codes.Internal, err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var item pipelineDM.Pipeline
		if err = r.db.ScanRows(rows, &item); err != nil {
			return nil, 0, "", status.Error(codes.Internal, err.Error())
		}
		createTime = item.CreateTime
		pipelines = append(pipelines, item)
	}

	if len(pipelines) > 0 {
		lastUID := (pipelines)[len(pipelines)-1].UID
		lastItem := &pipelineDM.Pipeline{}
		if result := r.db.Model(&pipelineDM.Pipeline{}).
			Where("owner = ?", owner).
			Order("create_time ASC, uid ASC").
			Limit(1).Find(lastItem); result.Error != nil {
			return nil, 0, "", status.Errorf(codes.Internal, result.Error.Error())
		}
		if lastItem.UID.String() == lastUID.String() {
			nextPageToken = ""
		} else {
			nextPageToken = paginate.EncodeToken(createTime, lastUID.String())
		}
	}

	return pipelines, totalSize, nextPageToken, nil
}

func (r *repository) GetPipelineByID(id string, owner string, isBasicView bool) (*pipelineDM.Pipeline, error) {
	queryBuilder := r.db.Model(&pipelineDM.Pipeline{}).Where("id = ? AND owner = ?", id, owner)
	if isBasicView {
		queryBuilder.Omit("pipeline.recipe")
	}
	var pipeline pipelineDM.Pipeline
	if result := queryBuilder.First(&pipeline); result.Error != nil {
		return nil, status.Errorf(codes.NotFound, "[GetPipelineByID] The pipeline id %s you specified is not found", id)
	}
	return &pipeline, nil
}

func (r *repository) GetPipelineByUID(uid uuid.UUID, owner string, isBasicView bool) (*pipelineDM.Pipeline, error) {
	queryBuilder := r.db.Model(&pipelineDM.Pipeline{}).Where("uid = ? AND owner = ?", uid, owner)
	if isBasicView {
		queryBuilder.Omit("pipeline.recipe")
	}
	var pipeline pipelineDM.Pipeline
	if result := queryBuilder.First(&pipeline); result.Error != nil {
		return nil, status.Errorf(codes.NotFound, "[GetPipelineByUID] The pipeline uid %s you specified is not found", uid.String())
	}
	return &pipeline, nil
}

// TranspileFilter transpiles a parsed AIP filter expression to GORM DB clauses
func (r *repository) transpileFilter(filter filtering.Filter) (*clause.Expr, error) {
	return (&Transpiler{
		filter: filter,
	}).Transpile()
}
