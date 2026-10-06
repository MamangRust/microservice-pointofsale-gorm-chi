package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"gorm.io/gorm"
)

type roleCommandRepository struct {
	db *gorm.DB
}

func NewRoleCommandRepository(db *gorm.DB) *roleCommandRepository {
	return &roleCommandRepository{db: db}
}

func timePtr(t time.Time) *time.Time { return &t }

func (r *roleCommandRepository) CreateRole(ctx context.Context, req *requests.CreateRoleRequest) (*models.Role, error) {
	role := &models.Role{
		RoleName:  req.Name,
		CreatedAt: timePtr(time.Now()),
		UpdatedAt: timePtr(time.Now()),
	}
	if err := r.db.WithContext(ctx).Create(role).Error; err != nil {
		return nil, err
	}
	return role, nil
}

func (r *roleCommandRepository) UpdateRole(ctx context.Context, req *requests.UpdateRoleRequest) (*models.Role, error) {
	var role models.Role
	if err := r.db.WithContext(ctx).First(&role, *req.ID).Error; err != nil {
		return nil, err
	}
	role.RoleName = req.Name
	role.UpdatedAt = timePtr(time.Now())
	if err := r.db.WithContext(ctx).Save(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleCommandRepository) TrashedRole(ctx context.Context, id int) (*models.Role, error) {
	var role models.Role
	if err := r.db.WithContext(ctx).First(&role, id).Error; err != nil {
		return nil, err
	}
	now := timePtr(time.Now())
	if err := r.db.WithContext(ctx).Model(&role).Update("deleted_at", now).Error; err != nil {
		return nil, err
	}
	role.DeletedAt = now
	return &role, nil
}

func (r *roleCommandRepository) RestoreRole(ctx context.Context, id int) (*models.Role, error) {
	var role models.Role
	if err := r.db.WithContext(ctx).Unscoped().Where("role_id = ?", id).First(&role).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Unscoped().Model(&role).Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}
	role.DeletedAt = nil
	return &role, nil
}

func (r *roleCommandRepository) DeleteRolePermanent(ctx context.Context, id int) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("role_id = ?", id).Delete(&models.Role{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *roleCommandRepository) RestoreAllRole(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Model(&models.Role{}).Update("deleted_at", nil)
	if result.Error != nil {
		return false, result.Error
	}
	return true, nil
}

func (r *roleCommandRepository) DeleteAllRolePermanent(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Role{})
	if result.Error != nil {
		return false, result.Error
	}
	return true, nil
}
