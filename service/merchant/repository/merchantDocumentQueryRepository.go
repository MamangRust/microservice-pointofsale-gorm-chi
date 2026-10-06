package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"gorm.io/gorm"
)

type merchantDocumentQueryRepository struct {
	db *gorm.DB
}

func NewMerchantDocumentQueryRepository(db *gorm.DB) MerchantDocumentQueryRepository {
	return &merchantDocumentQueryRepository{db: db}
}

func (r *merchantDocumentQueryRepository) FindAllDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*MerchantDocumentResult, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*MerchantDocumentResult

	err := r.db.WithContext(ctx).Raw(`
		SELECT d.document_id, d.merchant_id, d.document_type, d.document_url, d.status, d.note,
			COALESCE(TO_CHAR(d.uploaded_at, 'YYYY-MM-DD HH24:MI:SS'), '') as uploaded_at,
			COALESCE(TO_CHAR(d.created_at, 'YYYY-MM-DD HH24:MI:SS'), '') as created_at,
			COALESCE(TO_CHAR(d.updated_at, 'YYYY-MM-DD HH24:MI:SS'), '') as updated_at,
			COUNT(*) OVER() as total_count
		FROM merchant_documents d
		WHERE d.deleted_at IS NULL
		AND (? = '' OR d.document_type ILIKE ? OR d.document_url ILIKE ?)
		ORDER BY d.document_id DESC
		LIMIT ? OFFSET ?
	`, req.Search, "%"+req.Search+"%", "%"+req.Search+"%", req.PageSize, offset).Scan(&results).Error
	if err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var totalCount int
	if len(results) > 0 {
		totalCount = int(results[0].TotalCount)
	}
	return results, &totalCount, nil
}

func (r *merchantDocumentQueryRepository) FindById(ctx context.Context, id int) (*models.MerchantDocument, error) {
	var doc models.MerchantDocument
	if err := r.db.WithContext(ctx).Where("document_id = ? AND deleted_at IS NULL", id).First(&doc).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*MerchantDocumentResultDeleteAt, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*MerchantDocumentResultDeleteAt

	err := r.db.WithContext(ctx).Raw(`
		SELECT d.document_id, d.merchant_id, d.document_type, d.document_url, d.status, d.note,
			COALESCE(TO_CHAR(d.uploaded_at, 'YYYY-MM-DD HH24:MI:SS'), '') as uploaded_at,
			COALESCE(TO_CHAR(d.created_at, 'YYYY-MM-DD HH24:MI:SS'), '') as created_at,
			COALESCE(TO_CHAR(d.updated_at, 'YYYY-MM-DD HH24:MI:SS'), '') as updated_at,
			COALESCE(TO_CHAR(d.deleted_at, 'YYYY-MM-DD HH24:MI:SS'), '') as deleted_at,
			COUNT(*) OVER() as total_count
		FROM merchant_documents d
		WHERE d.deleted_at IS NULL
		AND (? = '' OR d.document_type ILIKE ? OR d.document_url ILIKE ?)
		ORDER BY d.document_id DESC
		LIMIT ? OFFSET ?
	`, req.Search, "%"+req.Search+"%", "%"+req.Search+"%", req.PageSize, offset).Scan(&results).Error
	if err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var totalCount int
	if len(results) > 0 {
		totalCount = int(results[0].TotalCount)
	}
	return results, &totalCount, nil
}

func (r *merchantDocumentQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*MerchantDocumentResultDeleteAt, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*MerchantDocumentResultDeleteAt

	err := r.db.WithContext(ctx).Raw(`
		SELECT d.document_id, d.merchant_id, d.document_type, d.document_url, d.status, d.note,
			COALESCE(TO_CHAR(d.uploaded_at, 'YYYY-MM-DD HH24:MI:SS'), '') as uploaded_at,
			COALESCE(TO_CHAR(d.created_at, 'YYYY-MM-DD HH24:MI:SS'), '') as created_at,
			COALESCE(TO_CHAR(d.updated_at, 'YYYY-MM-DD HH24:MI:SS'), '') as updated_at,
			COALESCE(TO_CHAR(d.deleted_at, 'YYYY-MM-DD HH24:MI:SS'), '') as deleted_at,
			COUNT(*) OVER() as total_count
		FROM merchant_documents d
		WHERE d.deleted_at IS NOT NULL
		AND (? = '' OR d.document_type ILIKE ? OR d.document_url ILIKE ?)
		ORDER BY d.document_id DESC
		LIMIT ? OFFSET ?
	`, req.Search, "%"+req.Search+"%", "%"+req.Search+"%", req.PageSize, offset).Scan(&results).Error
	if err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var totalCount int
	if len(results) > 0 {
		totalCount = int(results[0].TotalCount)
	}
	return results, &totalCount, nil
}
