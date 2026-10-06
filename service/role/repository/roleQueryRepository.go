package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"gorm.io/gorm"
)

type roleQueryRepository struct {
	db *gorm.DB
}

func NewRoleQueryRepository(db *gorm.DB) *roleQueryRepository {
	return &roleQueryRepository{db: db}
}

func (r *roleQueryRepository) FindAllRoles(ctx context.Context, req *requests.FindAllRoles) ([]*RoleResult, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*RoleResult
	err := r.db.WithContext(ctx).Raw(`
		SELECT r.role_id, r.role_name,
			COALESCE(TO_CHAR(r.created_at, 'YYYY-MM-DD HH24:MI:SS'), '') as created_at,
			COALESCE(TO_CHAR(r.updated_at, 'YYYY-MM-DD HH24:MI:SS'), '') as updated_at,
			COALESCE(TO_CHAR(r.deleted_at, 'YYYY-MM-DD HH24:MI:SS'), '') as deleted_at,
			COUNT(*) OVER() as total_count
		FROM roles r
		WHERE r.deleted_at IS NULL
			AND (? = '' OR r.role_name ILIKE ?)
		ORDER BY r.role_id DESC
		LIMIT ? OFFSET ?
	`, req.Search, "%"+req.Search+"%", req.PageSize, offset).Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *roleQueryRepository) FindById(ctx context.Context, id int) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Where("role_id = ? AND deleted_at IS NULL", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleQueryRepository) FindByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Where("role_name = ? AND deleted_at IS NULL", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleQueryRepository) FindByUserId(ctx context.Context, userID int) ([]*models.Role, error) {
	var roles []*models.Role
	err := r.db.WithContext(ctx).Raw(`
		SELECT r.* FROM roles r
		INNER JOIN user_roles ur ON ur.role_id = r.role_id
		WHERE ur.user_id = ? AND r.deleted_at IS NULL AND ur.deleted_at IS NULL
	`, userID).Scan(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleQueryRepository) FindByActiveRole(ctx context.Context, req *requests.FindAllRoles) ([]*RoleResult, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*RoleResult
	err := r.db.WithContext(ctx).Raw(`
		SELECT r.role_id, r.role_name,
			COALESCE(TO_CHAR(r.created_at, 'YYYY-MM-DD HH24:MI:SS'), '') as created_at,
			COALESCE(TO_CHAR(r.updated_at, 'YYYY-MM-DD HH24:MI:SS'), '') as updated_at,
			COALESCE(TO_CHAR(r.deleted_at, 'YYYY-MM-DD HH24:MI:SS'), '') as deleted_at,
			COUNT(*) OVER() as total_count
		FROM roles r
		WHERE r.deleted_at IS NULL
			AND (? = '' OR r.role_name ILIKE ?)
		ORDER BY r.role_id DESC
		LIMIT ? OFFSET ?
	`, req.Search, "%"+req.Search+"%", req.PageSize, offset).Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *roleQueryRepository) FindByTrashedRole(ctx context.Context, req *requests.FindAllRoles) ([]*RoleResult, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*RoleResult
	err := r.db.WithContext(ctx).Raw(`
		SELECT r.role_id, r.role_name,
			COALESCE(TO_CHAR(r.created_at, 'YYYY-MM-DD HH24:MI:SS'), '') as created_at,
			COALESCE(TO_CHAR(r.updated_at, 'YYYY-MM-DD HH24:MI:SS'), '') as updated_at,
			COALESCE(TO_CHAR(r.deleted_at, 'YYYY-MM-DD HH24:MI:SS'), '') as deleted_at,
			COUNT(*) OVER() as total_count
		FROM roles r
		WHERE r.deleted_at IS NOT NULL
			AND (? = '' OR r.role_name ILIKE ?)
		ORDER BY r.role_id DESC
		LIMIT ? OFFSET ?
	`, req.Search, "%"+req.Search+"%", req.PageSize, offset).Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}
