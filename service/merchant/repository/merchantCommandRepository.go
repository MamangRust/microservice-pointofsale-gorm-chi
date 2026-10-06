package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"gorm.io/gorm"
)

type merchantCommandRepository struct {
	db *gorm.DB
}

func NewMerchantCommandRepository(db *gorm.DB) MerchantCommandRepository {
	return &merchantCommandRepository{db: db}
}

func (r *merchantCommandRepository) CreateMerchant(ctx context.Context, request *requests.CreateMerchantRequest) (*models.Merchant, error) {
	now := time.Now()
	merchant := &models.Merchant{
		UserID:       int32(request.UserID),
		Name:         request.Name,
		Description:  &request.Description,
		Address:      &request.Address,
		ContactEmail: &request.ContactEmail,
		ContactPhone: &request.ContactPhone,
		Status:       "inactive",
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}
	if err := r.db.WithContext(ctx).Create(merchant).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return merchant, nil
}

func (r *merchantCommandRepository) UpdateMerchant(ctx context.Context, request *requests.UpdateMerchantRequest) (*models.Merchant, error) {
	var merchant models.Merchant
	if err := r.db.WithContext(ctx).Where("merchant_id = ?", *request.MerchantID).First(&merchant).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	merchant.Name = request.Name
	merchant.Description = &request.Description
	merchant.Address = &request.Address
	merchant.ContactEmail = &request.ContactEmail
	merchant.ContactPhone = &request.ContactPhone
	merchant.Status = request.Status
	merchant.UpdatedAt = timePtr(time.Now())

	if err := r.db.WithContext(ctx).Save(&merchant).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &merchant, nil
}

func (r *merchantCommandRepository) UpdateMerchantStatus(ctx context.Context, request *requests.UpdateMerchantStatusRequest) (*models.Merchant, error) {
	var merchant models.Merchant
	if err := r.db.WithContext(ctx).Where("merchant_id = ?", *request.MerchantID).First(&merchant).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	merchant.Status = request.Status
	merchant.UpdatedAt = timePtr(time.Now())

	if err := r.db.WithContext(ctx).Save(&merchant).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &merchant, nil
}

func (r *merchantCommandRepository) TrashedMerchant(ctx context.Context, merchantID int) (*models.Merchant, error) {
	var merchant models.Merchant
	if err := r.db.WithContext(ctx).Where("merchant_id = ? AND deleted_at IS NULL", merchantID).First(&merchant).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	now := time.Now()
	merchant.DeletedAt = &now
	if err := r.db.WithContext(ctx).Save(&merchant).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &merchant, nil
}

func (r *merchantCommandRepository) RestoreMerchant(ctx context.Context, merchantID int) (*models.Merchant, error) {
	var merchant models.Merchant
	if err := r.db.WithContext(ctx).Unscoped().Where("merchant_id = ? AND deleted_at IS NOT NULL", merchantID).First(&merchant).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	merchant.DeletedAt = nil
	merchant.UpdatedAt = timePtr(time.Now())
	if err := r.db.WithContext(ctx).Unscoped().Save(&merchant).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &merchant, nil
}

func (r *merchantCommandRepository) DeleteMerchantPermanent(ctx context.Context, merchantID int) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("merchant_id = ? AND deleted_at IS NOT NULL", merchantID).Delete(&models.Merchant{})
	if result.Error != nil {
		return false, sharedErrors.ErrInternal.WithInternal(result.Error)
	}
	return result.RowsAffected > 0, nil
}

func (r *merchantCommandRepository) RestoreAllMerchant(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Model(&models.Merchant{}).Update("deleted_at", nil)
	if result.Error != nil {
		return false, sharedErrors.ErrInternal.WithInternal(result.Error)
	}
	return true, nil
}

func (r *merchantCommandRepository) DeleteAllMerchantPermanent(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Merchant{})
	if result.Error != nil {
		return false, sharedErrors.ErrInternal.WithInternal(result.Error)
	}
	return true, nil
}

func timePtr(t time.Time) *time.Time { return &t }
