package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"gorm.io/gorm"
)

type resetTokenRepository struct {
	db *gorm.DB
}

func NewResetTokenRepository(db *gorm.DB) *resetTokenRepository {
	return &resetTokenRepository{db: db}
}

func (r *resetTokenRepository) FindByToken(ctx context.Context, code string) (*models.ResetToken, error) {
	var rt models.ResetToken
	err := r.db.WithContext(ctx).Where("token = ?", code).First(&rt).Error
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &rt, nil
}

func (r *resetTokenRepository) CreateResetToken(ctx context.Context, req *requests.CreateResetTokenRequest) (*models.ResetToken, error) {
	expiryDate, err := time.Parse("2006-01-02 15:04:05", req.ExpiredAt)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	rt := &models.ResetToken{
		UserID:     int64(req.UserID),
		Token:      req.ResetToken,
		ExpiryDate: expiryDate,
	}
	if err := r.db.WithContext(ctx).Create(rt).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return rt, nil
}

func (r *resetTokenRepository) DeleteResetToken(ctx context.Context, user_id int) error {
	err := r.db.WithContext(ctx).Where("user_id = ?", user_id).Delete(&models.ResetToken{}).Error
	if err != nil {
		return sharedErrors.ErrInternal.WithInternal(err)
	}
	return nil
}
