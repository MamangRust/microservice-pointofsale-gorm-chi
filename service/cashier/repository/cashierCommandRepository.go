package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/cashier_errors"
	"gorm.io/gorm"
)

type cashierCommandRepository struct {
	db *gorm.DB
}

func NewCashierCommandRepository(db *gorm.DB) CashierCommandRepository {
	return &cashierCommandRepository{db: db}
}

func (r *cashierCommandRepository) CreateCashier(ctx context.Context, request *requests.CreateCashierRequest) (*models.Cashier, error) {
	cashier := &models.Cashier{
		MerchantID: int32(request.MerchantID),
		UserID:     int32(request.UserID),
		Name:       request.Name,
	}
	if err := r.db.WithContext(ctx).Create(cashier).Error; err != nil {
		return nil, cashier_errors.ErrCreateCashier
	}
	return cashier, nil
}

func (r *cashierCommandRepository) UpdateCashier(ctx context.Context, request *requests.UpdateCashierRequest) (*models.Cashier, error) {
	var cashier models.Cashier
	if err := r.db.WithContext(ctx).Where("cashier_id = ?", *request.CashierID).First(&cashier).Error; err != nil {
		return nil, cashier_errors.ErrUpdateCashier
	}
	cashier.Name = request.Name
	if err := r.db.WithContext(ctx).Save(&cashier).Error; err != nil {
		return nil, cashier_errors.ErrUpdateCashier
	}
	return &cashier, nil
}

func (r *cashierCommandRepository) TrashedCashier(ctx context.Context, cashier_id int) (*models.Cashier, error) {
	var cashier models.Cashier
	if err := r.db.WithContext(ctx).Where("cashier_id = ?", cashier_id).First(&cashier).Error; err != nil {
		return nil, cashier_errors.ErrTrashedCashier
	}
	tm := time.Now()
	if err := r.db.WithContext(ctx).Model(&cashier).Update("deleted_at", tm).Error; err != nil {
		return nil, cashier_errors.ErrTrashedCashier
	}
	return &cashier, nil
}

func (r *cashierCommandRepository) RestoreCashier(ctx context.Context, cashier_id int) (*models.Cashier, error) {
	var cashier models.Cashier
	if err := r.db.WithContext(ctx).Unscoped().Where("cashier_id = ?", cashier_id).First(&cashier).Error; err != nil {
		return nil, cashier_errors.ErrRestoreCashier
	}
	if err := r.db.WithContext(ctx).Unscoped().Model(&cashier).Update("deleted_at", nil).Error; err != nil {
		return nil, cashier_errors.ErrRestoreCashier
	}
	return &cashier, nil
}

func (r *cashierCommandRepository) DeleteCashierPermanent(ctx context.Context, cashier_id int) (bool, error) {
	if err := r.db.WithContext(ctx).Unscoped().Where("cashier_id = ?", cashier_id).Delete(&models.Cashier{}).Error; err != nil {
		return false, cashier_errors.ErrDeleteCashierPermanent
	}
	return true, nil
}

func (r *cashierCommandRepository) RestoreAllCashier(ctx context.Context) (bool, error) {
	if err := r.db.WithContext(ctx).Unscoped().Model(&models.Cashier{}).Where("deleted_at IS NOT NULL").Update("deleted_at", nil).Error; err != nil {
		return false, cashier_errors.ErrRestoreAllCashiers
	}
	return true, nil
}

func (r *cashierCommandRepository) DeleteAllCashierPermanent(ctx context.Context) (bool, error) {
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Cashier{}).Error; err != nil {
		return false, cashier_errors.ErrDeleteAllCashiersPermanent
	}
	return true, nil
}
