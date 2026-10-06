package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/transaction_errors"
	"gorm.io/gorm"
)

type transactionCommandRepository struct {
	db *gorm.DB
}

func NewTransactionCommandRepository(db *gorm.DB) *transactionCommandRepository {
	return &transactionCommandRepository{db: db}
}

func (r *transactionCommandRepository) CreateTransaction(ctx context.Context, request *requests.CreateTransactionRequest) (*models.Transaction, error) {
	now := time.Now()
	t := &models.Transaction{
		OrderID:       int32(request.OrderID),
		MerchantID:    int32(request.MerchantID),
		PaymentMethod: request.PaymentMethod,
		Amount:        int32(request.Amount),
		PaymentStatus: request.PaymentStatus,
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}

	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return nil, transaction_errors.ErrCreateTransaction
	}
	return t, nil
}

func (r *transactionCommandRepository) UpdateTransaction(ctx context.Context, request *requests.UpdateTransactionRequest) (*models.Transaction, error) {
	var t models.Transaction
	if err := r.db.WithContext(ctx).Where("transaction_id = ? AND deleted_at IS NULL", *request.TransactionID).First(&t).Error; err != nil {
		return nil, transaction_errors.ErrUpdateTransaction
	}

	now := time.Now()
	t.MerchantID = int32(request.MerchantID)
	t.PaymentMethod = request.PaymentMethod
	t.Amount = int32(request.Amount)
	t.OrderID = int32(request.OrderID)
	t.PaymentStatus = request.PaymentStatus
	t.UpdatedAt = &now

	if err := r.db.WithContext(ctx).Save(&t).Error; err != nil {
		return nil, transaction_errors.ErrUpdateTransaction
	}
	return &t, nil
}

func (r *transactionCommandRepository) TrashTransaction(ctx context.Context, transaction_id int) (*models.Transaction, error) {
	var t models.Transaction
	if err := r.db.WithContext(ctx).Where("transaction_id = ? AND deleted_at IS NULL", transaction_id).First(&t).Error; err != nil {
		return nil, transaction_errors.ErrTrashTransaction
	}

	now := time.Now()
	t.DeletedAt = &now

	if err := r.db.WithContext(ctx).Save(&t).Error; err != nil {
		return nil, transaction_errors.ErrTrashTransaction
	}
	return &t, nil
}

func (r *transactionCommandRepository) RestoreTransaction(ctx context.Context, transaction_id int) (*models.Transaction, error) {
	var t models.Transaction
	if err := r.db.WithContext(ctx).Unscoped().Where("transaction_id = ? AND deleted_at IS NOT NULL", transaction_id).First(&t).Error; err != nil {
		return nil, transaction_errors.ErrRestoreTransaction
	}

	t.DeletedAt = nil
	if err := r.db.WithContext(ctx).Unscoped().Save(&t).Error; err != nil {
		return nil, transaction_errors.ErrRestoreTransaction
	}
	return &t, nil
}

func (r *transactionCommandRepository) DeleteTransactionPermanently(ctx context.Context, transaction_id int) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("transaction_id = ?", transaction_id).Delete(&models.Transaction{})
	if result.Error != nil {
		return false, transaction_errors.ErrDeleteTransactionPermanently
	}
	if result.RowsAffected == 0 {
		return false, transaction_errors.ErrDeleteTransactionPermanently
	}
	return true, nil
}

func (r *transactionCommandRepository) RestoreAllTransactions(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Model(&models.Transaction{}).Update("deleted_at", nil)
	if result.Error != nil {
		return false, transaction_errors.ErrRestoreAllTransactions
	}
	return true, nil
}

func (r *transactionCommandRepository) DeleteAllTransactionPermanent(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Transaction{})
	if result.Error != nil {
		return false, transaction_errors.ErrDeleteAllTransactionPermanent
	}
	return true, nil
}
