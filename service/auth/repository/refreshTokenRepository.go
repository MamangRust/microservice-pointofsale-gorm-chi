package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"gorm.io/gorm"
)

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *refreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) FindByToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&rt).Error
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &rt, nil
}

func (r *refreshTokenRepository) FindByUserId(ctx context.Context, user_id int) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.db.WithContext(ctx).Where("user_id = ?", user_id).First(&rt).Error
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &rt, nil
}

func (r *refreshTokenRepository) CreateRefreshToken(ctx context.Context, req *requests.CreateRefreshToken) (*models.RefreshToken, error) {
	layout := "2006-01-02 15:04:05"
	expirationTime, err := time.Parse(layout, req.ExpiresAt)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	rt := &models.RefreshToken{
		UserID:     int32(req.UserId),
		Token:      req.Token,
		Expiration: expirationTime,
	}
	if err := r.db.WithContext(ctx).Create(rt).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return rt, nil
}

func (r *refreshTokenRepository) UpdateRefreshToken(ctx context.Context, req *requests.UpdateRefreshToken) (*models.RefreshToken, error) {
	layout := "2006-01-02 15:04:05"
	expirationTime, err := time.Parse(layout, req.ExpiresAt)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var rt models.RefreshToken
	err = r.db.WithContext(ctx).Where("user_id = ?", req.UserId).First(&rt).Error
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	rt.Token = req.Token
	rt.Expiration = expirationTime
	if err := r.db.WithContext(ctx).Save(&rt).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &rt, nil
}

func (r *refreshTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	err := r.db.WithContext(ctx).Where("token = ?", token).Delete(&models.RefreshToken{}).Error
	if err != nil {
		return sharedErrors.ErrInternal.WithInternal(err)
	}
	return nil
}

func (r *refreshTokenRepository) DeleteRefreshTokenByUserId(ctx context.Context, user_id int) error {
	err := r.db.WithContext(ctx).Where("user_id = ?", user_id).Delete(&models.RefreshToken{}).Error
	if err != nil {
		return sharedErrors.ErrInternal.WithInternal(err)
	}
	return nil
}
