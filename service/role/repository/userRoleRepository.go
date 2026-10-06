package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"gorm.io/gorm"
)

type userRoleRepository struct {
	db *gorm.DB
}

func NewUserRoleRepository(db *gorm.DB) *userRoleRepository {
	return &userRoleRepository{db: db}
}

func (r *userRoleRepository) AssignRoleToUser(ctx context.Context, req *requests.CreateUserRoleRequest) (*models.UserRole, error) {
	ur := &models.UserRole{
		UserID: int32(req.UserId),
		RoleID: int32(req.RoleId),
	}
	if err := r.db.WithContext(ctx).Create(ur).Error; err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return ur, nil
}

func (r *userRoleRepository) RemoveRoleFromUser(ctx context.Context, req *requests.RemoveUserRoleRequest) error {
	err := r.db.WithContext(ctx).Where("user_id = ? AND role_id = ?", req.UserId, req.RoleId).Delete(&models.UserRole{}).Error
	if err != nil {
		return sharedErrors.ErrInternal.WithInternal(err)
	}
	return nil
}
