package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"gorm.io/gorm"
)

type userCommandRepository struct {
	db *gorm.DB
}

func NewUserCommandRepository(db *gorm.DB) *userCommandRepository {
	return &userCommandRepository{db: db}
}

func timePtr(t time.Time) *time.Time { return &t }

func (r *userCommandRepository) CreateUser(ctx context.Context, request *requests.CreateUserRequest) (*models.User, error) {
	user := &models.User{
		Firstname: request.FirstName,
		Lastname:  request.LastName,
		Email:     request.Email,
		Password:  request.Password,
		CreatedAt: timePtr(time.Now()),
		UpdatedAt: timePtr(time.Now()),
	}
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// CreateUserRecord registers a user with auth-managed fields. The password is
// expected to be already hashed by the auth service; verification fields pass
// through untouched.
func (r *userCommandRepository) CreateUserRecord(ctx context.Context, request *requests.RegisterRequest) (*models.User, error) {
	user := &models.User{
		Firstname:        request.FirstName,
		Lastname:         request.LastName,
		Email:            request.Email,
		Password:         request.Password,
		VerificationCode: request.VerifiedCode,
		IsVerified:       &request.IsVerified,
		CreatedAt:        timePtr(time.Now()),
		UpdatedAt:        timePtr(time.Now()),
	}
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userCommandRepository) UpdateUserIsVerified(ctx context.Context, userID int, isVerified bool) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		return nil, err
	}
	user.IsVerified = &isVerified
	user.UpdatedAt = timePtr(time.Now())
	if err := r.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userCommandRepository) UpdateUserPassword(ctx context.Context, userID int, password string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		return nil, err
	}
	user.Password = password
	user.UpdatedAt = timePtr(time.Now())
	if err := r.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userCommandRepository) UpdateUser(ctx context.Context, request *requests.UpdateUserRequest) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, *request.UserID).Error; err != nil {
		return nil, err
	}
	user.Firstname = request.FirstName
	user.Lastname = request.LastName
	user.Email = request.Email
	user.Password = request.Password
	user.UpdatedAt = timePtr(time.Now())
	if err := r.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userCommandRepository) TrashedUser(ctx context.Context, userID int) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}
	now := timePtr(time.Now())
	if err := r.db.WithContext(ctx).Model(&user).Update("deleted_at", now).Error; err != nil {
		return nil, err
	}
	user.DeletedAt = now
	return &user, nil
}

func (r *userCommandRepository) RestoreUser(ctx context.Context, userID int) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Unscoped().Where("user_id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Unscoped().Model(&user).Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}
	user.DeletedAt = nil
	return &user, nil
}

func (r *userCommandRepository) DeleteUserPermanent(ctx context.Context, userID int) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("user_id = ?", userID).Delete(&models.User{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *userCommandRepository) RestoreAllUser(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Model(&models.User{}).Update("deleted_at", nil)
	return result.Error == nil, result.Error
}

func (r *userCommandRepository) DeleteAllUserPermanent(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.User{})
	return result.Error == nil, result.Error
}
