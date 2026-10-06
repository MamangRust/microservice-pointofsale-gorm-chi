package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/transaction_errors"
	"gorm.io/gorm"
)

type transactionQueryRepository struct {
	db *gorm.DB
}

func NewTransactionQueryRepository(db *gorm.DB) *transactionQueryRepository {
	return &transactionQueryRepository{db: db}
}

func (r *transactionQueryRepository) FindAllTransactions(ctx context.Context, req *requests.FindAllTransaction) ([]*TransactionResult, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	type row struct {
		TransactionID int32
		OrderID       int32
		MerchantID    int32
		PaymentMethod string
		Amount        int32
		ChangeAmount  *int32
		PaymentStatus string
		CreatedAt     string
		TotalCount    int64
	}

	var results []row
	query := r.db.WithContext(ctx).Raw(`
		SELECT t.transaction_id, t.order_id, t.merchant_id, t.payment_method, t.amount,
		       t.change_amount, t.payment_status,
		       TO_CHAR(t.created_at, 'YYYY-MM-DD HH24:MI:SS') AS created_at,
		       COUNT(*) OVER() AS total_count
		FROM transactions t
		WHERE t.deleted_at IS NULL
		  AND (t.payment_method ILIKE ? OR ? = '')
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?
	`, "%"+req.Search+"%", req.Search, req.PageSize, offset)

	if err := query.Scan(&results).Error; err != nil {
		return nil, nil, transaction_errors.ErrFindAllTransactions
	}

	var totalCount int
	if len(results) > 0 {
		totalCount = int(results[0].TotalCount)
	}

	var out []*TransactionResult
	for _, r := range results {
		cr := r
		out = append(out, &TransactionResult{
			TransactionID: cr.TransactionID, OrderID: cr.OrderID, MerchantID: cr.MerchantID,
			PaymentMethod: cr.PaymentMethod, Amount: cr.Amount, ChangeAmount: cr.ChangeAmount,
			PaymentStatus: cr.PaymentStatus, CreatedAt: cr.CreatedAt, TotalCount: cr.TotalCount,
		})
	}
	return out, &totalCount, nil
}

func (r *transactionQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllTransaction) ([]*TransactionResultDeleteAt, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	type row struct {
		TransactionID int32
		OrderID       int32
		MerchantID    int32
		PaymentMethod string
		Amount        int32
		ChangeAmount  *int32
		PaymentStatus string
		CreatedAt     string
		DeletedAt     string
		TotalCount    int64
	}

	var results []row
	query := r.db.WithContext(ctx).Raw(`
		SELECT t.transaction_id, t.order_id, t.merchant_id, t.payment_method, t.amount,
		       t.change_amount, t.payment_status,
		       TO_CHAR(t.created_at, 'YYYY-MM-DD HH24:MI:SS') AS created_at,
		       COALESCE(TO_CHAR(t.deleted_at, 'YYYY-MM-DD HH24:MI:SS'), '') AS deleted_at,
		       COUNT(*) OVER() AS total_count
		FROM transactions t
		WHERE t.deleted_at IS NULL
		  AND (t.payment_method ILIKE ? OR ? = '')
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?
	`, "%"+req.Search+"%", req.Search, req.PageSize, offset)

	if err := query.Scan(&results).Error; err != nil {
		return nil, nil, transaction_errors.ErrFindByActive
	}

	var totalCount int
	if len(results) > 0 {
		totalCount = int(results[0].TotalCount)
	}

	var out []*TransactionResultDeleteAt
	for _, r := range results {
		cr := r
		out = append(out, &TransactionResultDeleteAt{
			TransactionID: cr.TransactionID, OrderID: cr.OrderID, MerchantID: cr.MerchantID,
			PaymentMethod: cr.PaymentMethod, Amount: cr.Amount, ChangeAmount: cr.ChangeAmount,
			PaymentStatus: cr.PaymentStatus, CreatedAt: cr.CreatedAt, DeletedAt: cr.DeletedAt,
			TotalCount: cr.TotalCount,
		})
	}
	return out, &totalCount, nil
}

func (r *transactionQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllTransaction) ([]*TransactionResultDeleteAt, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	type row struct {
		TransactionID int32
		OrderID       int32
		MerchantID    int32
		PaymentMethod string
		Amount        int32
		ChangeAmount  *int32
		PaymentStatus string
		CreatedAt     string
		DeletedAt     string
		TotalCount    int64
	}

	var results []row
	query := r.db.WithContext(ctx).Raw(`
		SELECT t.transaction_id, t.order_id, t.merchant_id, t.payment_method, t.amount,
		       t.change_amount, t.payment_status,
		       TO_CHAR(t.created_at, 'YYYY-MM-DD HH24:MI:SS') AS created_at,
		       COALESCE(TO_CHAR(t.deleted_at, 'YYYY-MM-DD HH24:MI:SS'), '') AS deleted_at,
		       COUNT(*) OVER() AS total_count
		FROM transactions t
		WHERE t.deleted_at IS NOT NULL
		  AND (t.payment_method ILIKE ? OR ? = '')
		ORDER BY t.deleted_at DESC
		LIMIT ? OFFSET ?
	`, "%"+req.Search+"%", req.Search, req.PageSize, offset)

	if err := query.Scan(&results).Error; err != nil {
		return nil, nil, transaction_errors.ErrFindByTrashed
	}

	var totalCount int
	if len(results) > 0 {
		totalCount = int(results[0].TotalCount)
	}

	var out []*TransactionResultDeleteAt
	for _, r := range results {
		cr := r
		out = append(out, &TransactionResultDeleteAt{
			TransactionID: cr.TransactionID, OrderID: cr.OrderID, MerchantID: cr.MerchantID,
			PaymentMethod: cr.PaymentMethod, Amount: cr.Amount, ChangeAmount: cr.ChangeAmount,
			PaymentStatus: cr.PaymentStatus, CreatedAt: cr.CreatedAt, DeletedAt: cr.DeletedAt,
			TotalCount: cr.TotalCount,
		})
	}
	return out, &totalCount, nil
}

func (r *transactionQueryRepository) FindByMerchant(ctx context.Context, req *requests.FindAllTransactionByMerchant) ([]*TransactionByMerchantResult, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	type row struct {
		TransactionID int32
		OrderID       int32
		MerchantID    int32
		PaymentMethod string
		Amount        int32
		ChangeAmount  *int32
		PaymentStatus string
		CreatedAt     string
		TotalCount    int64
	}

	var results []row
	query := r.db.WithContext(ctx).Raw(`
		SELECT t.transaction_id, t.order_id, t.merchant_id, t.payment_method, t.amount,
		       t.change_amount, t.payment_status,
		       TO_CHAR(t.created_at, 'YYYY-MM-DD HH24:MI:SS') AS created_at,
		       COUNT(*) OVER() AS total_count
		FROM transactions t
		WHERE t.deleted_at IS NULL
		  AND t.merchant_id = ?
		  AND (t.payment_method ILIKE ? OR ? = '')
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?
	`, req.MerchantID, "%"+req.Search+"%", req.Search, req.PageSize, offset)

	if err := query.Scan(&results).Error; err != nil {
		return nil, nil, transaction_errors.ErrFindByMerchant
	}

	var totalCount int
	if len(results) > 0 {
		totalCount = int(results[0].TotalCount)
	}

	var out []*TransactionByMerchantResult
	for _, r := range results {
		cr := r
		out = append(out, &TransactionByMerchantResult{
			TransactionID: cr.TransactionID, OrderID: cr.OrderID, MerchantID: cr.MerchantID,
			PaymentMethod: cr.PaymentMethod, Amount: cr.Amount, ChangeAmount: cr.ChangeAmount,
			PaymentStatus: cr.PaymentStatus, CreatedAt: cr.CreatedAt, TotalCount: cr.TotalCount,
		})
	}
	return out, &totalCount, nil
}

func (r *transactionQueryRepository) FindById(ctx context.Context, transaction_id int) (*models.Transaction, error) {
	var t models.Transaction
	if err := r.db.WithContext(ctx).Where("transaction_id = ? AND deleted_at IS NULL", transaction_id).First(&t).Error; err != nil {
		return nil, transaction_errors.ErrFindById
	}
	return &t, nil
}

func (r *transactionQueryRepository) FindByOrderId(ctx context.Context, order_id int) (*models.Transaction, error) {
	var t models.Transaction
	if err := r.db.WithContext(ctx).Where("order_id = ? AND deleted_at IS NULL", order_id).First(&t).Error; err != nil {
		return nil, transaction_errors.ErrFindByOrderId
	}
	return &t, nil
}

// fmtTxTime formats a *time.Time to a string; returns "" if nil.
func fmtTxTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// parseTxTime converts a gRPC timestamp string to *time.Time.
func parseTxTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return nil
		}
	}
	return &t
}

// fmtTxStr dereferences a *string; returns "" if nil.
func fmtTxStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// fmtTxIntPtr dereferences a *int32; returns 0 if nil.
func fmtTxIntPtr(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

// ptrInt32 returns a *int32 from an int32.
func ptrInt32(v int32) *int32 {
	return &v
}

// strPtr returns a *string from a string.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// timePtr parses a string time to *time.Time.
func timePtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return nil
		}
	}
	return &t
}

// parseTxNullableTime parses a GORM nullable time string into a string.
func parseTxNullableTime(s string) string {
	return fmt.Sprintf("%s", s)
}
