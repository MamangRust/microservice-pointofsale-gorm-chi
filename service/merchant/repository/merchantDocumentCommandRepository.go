package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"gorm.io/gorm"
)

type merchantDocumentCommandRepository struct {
	db *gorm.DB
}

func NewMerchantDocumentCommandRepository(db *gorm.DB) MerchantDocumentCommandRepository {
	return &merchantDocumentCommandRepository{db: db}
}

func (r *merchantDocumentCommandRepository) CreateMerchantDocument(ctx context.Context, request *requests.CreateMerchantDocumentRequest) (*models.MerchantDocument, error) {
	now := time.Now()
	emptyNote := ""
	doc := &models.MerchantDocument{
		MerchantID:   int32(request.MerchantID),
		DocumentType: request.DocumentType,
		DocumentUrl:  request.DocumentUrl,
		Status:       "pending",
		Note:         &emptyNote,
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}
	if err := r.db.WithContext(ctx).Create(doc).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return doc, nil
}

func (r *merchantDocumentCommandRepository) UpdateMerchantDocument(ctx context.Context, request *requests.UpdateMerchantDocumentRequest) (*models.MerchantDocument, error) {
	if request.DocumentID == nil {
		return nil, sharedErrors.ErrBadRequest.WithMessage("document id is required")
	}
	var doc models.MerchantDocument
	if err := r.db.WithContext(ctx).Where("document_id = ?", *request.DocumentID).First(&doc).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	doc.DocumentType = request.DocumentType
	doc.DocumentUrl = request.DocumentUrl
	doc.Status = request.Status
	doc.Note = &request.Note
	doc.UpdatedAt = timePtr(time.Now())

	if err := r.db.WithContext(ctx).Save(&doc).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentCommandRepository) UpdateMerchantDocumentStatus(ctx context.Context, request *requests.UpdateMerchantDocumentStatusRequest) (*models.MerchantDocument, error) {
	if request.DocumentID == nil {
		return nil, sharedErrors.ErrBadRequest.WithMessage("document id is required")
	}
	var doc models.MerchantDocument
	if err := r.db.WithContext(ctx).Where("document_id = ?", *request.DocumentID).First(&doc).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	doc.Status = request.Status
	doc.Note = &request.Note
	doc.UpdatedAt = timePtr(time.Now())

	if err := r.db.WithContext(ctx).Save(&doc).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentCommandRepository) TrashedMerchantDocument(ctx context.Context, documentID int) (*models.MerchantDocument, error) {
	var doc models.MerchantDocument
	if err := r.db.WithContext(ctx).Where("document_id = ? AND deleted_at IS NULL", documentID).First(&doc).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	now := time.Now()
	doc.DeletedAt = &now
	if err := r.db.WithContext(ctx).Save(&doc).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentCommandRepository) RestoreMerchantDocument(ctx context.Context, documentID int) (*models.MerchantDocument, error) {
	var doc models.MerchantDocument
	if err := r.db.WithContext(ctx).Unscoped().Where("document_id = ? AND deleted_at IS NOT NULL", documentID).First(&doc).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	doc.DeletedAt = nil
	doc.UpdatedAt = timePtr(time.Now())
	if err := r.db.WithContext(ctx).Unscoped().Save(&doc).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentCommandRepository) DeleteMerchantDocumentPermanent(ctx context.Context, documentID int) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("document_id = ? AND deleted_at IS NOT NULL", documentID).Delete(&models.MerchantDocument{})
	if result.Error != nil {
		return false, sharedErrors.ErrInternal.WithInternal(result.Error)
	}
	return result.RowsAffected > 0, nil
}

func (r *merchantDocumentCommandRepository) RestoreAllMerchantDocument(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Model(&models.MerchantDocument{}).Update("deleted_at", nil)
	if result.Error != nil {
		return false, sharedErrors.ErrInternal.WithInternal(result.Error)
	}
	return true, nil
}

func (r *merchantDocumentCommandRepository) DeleteAllMerchantDocumentPermanent(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.MerchantDocument{})
	if result.Error != nil {
		return false, sharedErrors.ErrInternal.WithInternal(result.Error)
	}
	return true, nil
}
