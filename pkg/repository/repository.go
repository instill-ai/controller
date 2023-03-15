package repository

// import (
// 	"errors"
// 	"time"

// 	"github.com/gofrs/uuid"
// 	"github.com/jackc/pgconn"
// 	"go.einride.tech/aip/filtering"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// 	"gorm.io/gorm"
// 	"gorm.io/gorm/clause"

// 	"github.com/instill-ai/x/paginate"
// 	modelDM "github.com/instill-ai/model-backend/pkg/datamodel"
// 	pipelineDM "github.com/instill-ai/pipeline-backend/pkg/datamodel"
// 	modelPB "github.com/instill-ai/protogen-go/vdp/model/v1alpha"
// )

// // DefaultPageSize is the default pagination page size when page size is not assigned
// const DefaultPageSize = 10

// // MaxPageSize is the maximum pagination page size if the assigned value is over this number
// const MaxPageSize = 100

// // Repository interface
// type Repository interface {
// 	GetModelById(owner string, modelID string, view modelPB.View) (modelDM.Model, error)
// 	GetModelByUid(owner string, modelUID uuid.UUID, view modelPB.View) (modelDM.Model, error)
// 	ListModelInstance(modelUID uuid.UUID, view modelPB.View, pageSize int, pageToken string) (instances []modelDM.ModelInstance, nextPageToken string, totalSize int64, err error)
// 	ListModel(owner string, view modelPB.View, pageSize int, pageToken string) (models []modelDM.Model, nextPageToken string, totalSize int64, err error)
// 	ListPipeline(owner string, pageSize int64, pageToken string, isBasicView bool, filter filtering.Filter) ([]pipelineDM.Pipeline, int64, string, error)
// 	GetPipelineByID(id string, owner string, isBasicView bool) (*pipelineDM.Pipeline, error)
// 	GetPipelineByUID(uid uuid.UUID, owner string, isBasicView bool) (*pipelineDM.Pipeline, error)
// }

// type repository struct {
// 	db *gorm.DB
// }

// // NewRepository initiates a repository instance
// func NewRepository(db *gorm.DB) Repository {
// 	return &repository{
// 		db: db,
// 	}
// }

// func (r *repository) ListPipeline(owner string, pageSize int64, pageToken string, isBasicView bool, filter filtering.Filter) (pipelines []pipelineDM.Pipeline, totalSize int64, nextPageToken string, err error) {

// 	if result := r.db.Model(&pipelineDM.Pipeline{}).Where("owner = ?", owner).Count(&totalSize); result.Error != nil {
// 		return nil, 0, "", status.Errorf(codes.Internal, result.Error.Error())
// 	}

// 	queryBuilder := r.db.Model(&pipelineDM.Pipeline{}).Order("create_time DESC, uid DESC").Where("owner = ?", owner)

// 	if pageSize == 0 {
// 		pageSize = DefaultPageSize
// 	} else if pageSize > MaxPageSize {
// 		pageSize = MaxPageSize
// 	}

// 	queryBuilder = queryBuilder.Limit(int(pageSize))

// 	if pageToken != "" {
// 		createTime, uid, err := paginate.DecodeToken(pageToken)
// 		if err != nil {
// 			return nil, 0, "", status.Errorf(codes.InvalidArgument, "Invalid page token: %s", err.Error())
// 		}
// 		queryBuilder = queryBuilder.Where("(create_time,uid) < (?::timestamp, ?)", createTime, uid)
// 	}

// 	if isBasicView {
// 		queryBuilder.Omit("pipeline.recipe")
// 	}

// 	if expr, err := r.transpileFilter(filter); err != nil {
// 		return nil, 0, "", status.Errorf(codes.Internal, err.Error())
// 	} else if expr != nil {
// 		queryBuilder.Clauses(expr)
// 	}

// 	var createTime time.Time
// 	rows, err := queryBuilder.Rows()
// 	if err != nil {
// 		return nil, 0, "", status.Errorf(codes.Internal, err.Error())
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var item pipelineDM.Pipeline
// 		if err = r.db.ScanRows(rows, &item); err != nil {
// 			return nil, 0, "", status.Error(codes.Internal, err.Error())
// 		}
// 		createTime = item.CreateTime
// 		pipelines = append(pipelines, item)
// 	}

// 	if len(pipelines) > 0 {
// 		lastUID := (pipelines)[len(pipelines)-1].UID
// 		lastItem := &pipelineDM.Pipeline{}
// 		if result := r.db.Model(&pipelineDM.Pipeline{}).
// 			Where("owner = ?", owner).
// 			Order("create_time ASC, uid ASC").
// 			Limit(1).Find(lastItem); result.Error != nil {
// 			return nil, 0, "", status.Errorf(codes.Internal, result.Error.Error())
// 		}
// 		if lastItem.UID.String() == lastUID.String() {
// 			nextPageToken = ""
// 		} else {
// 			nextPageToken = paginate.EncodeToken(createTime, lastUID.String())
// 		}
// 	}

// 	return pipelines, totalSize, nextPageToken, nil
// }

// func (r *repository) GetPipelineByID(id string, owner string, isBasicView bool) (*pipelineDM.Pipeline, error) {
// 	queryBuilder := r.db.Model(&pipelineDM.Pipeline{}).Where("id = ? AND owner = ?", id, owner)
// 	if isBasicView {
// 		queryBuilder.Omit("pipeline.recipe")
// 	}
// 	var pipeline pipelineDM.Pipeline
// 	if result := queryBuilder.First(&pipeline); result.Error != nil {
// 		return nil, status.Errorf(codes.NotFound, "[GetPipelineByID] The pipeline id %s you specified is not found", id)
// 	}
// 	return &pipeline, nil
// }

// func (r *repository) GetPipelineByUID(uid uuid.UUID, owner string, isBasicView bool) (*pipelineDM.Pipeline, error) {
// 	queryBuilder := r.db.Model(&pipelineDM.Pipeline{}).Where("uid = ? AND owner = ?", uid, owner)
// 	if isBasicView {
// 		queryBuilder.Omit("pipeline.recipe")
// 	}
// 	var pipeline pipelineDM.Pipeline
// 	if result := queryBuilder.First(&pipeline); result.Error != nil {
// 		return nil, status.Errorf(codes.NotFound, "[GetPipelineByUID] The pipeline uid %s you specified is not found", uid.String())
// 	}
// 	return &pipeline, nil
// }

// // TranspileFilter transpiles a parsed AIP filter expression to GORM DB clauses
// func (r *repository) transpileFilter(filter filtering.Filter) (*clause.Expr, error) {
// 	return (&Transpiler{
// 		filter: filter,
// 	}).Transpile()
// }