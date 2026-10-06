package transaction_test

import (
	"context"
	"testing"

	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	trans_cache "github.com/MamangRust/microservice-point-of-sale-transacton/cache"
	"github.com/MamangRust/microservice-point-of-sale-transacton/repository"
	"github.com/MamangRust/microservice-point-of-sale-transacton/service"
	"github.com/stretchr/testify/suite"
)

type TransactionServiceTestSuite struct {
	tests.BaseTestSuite
	svc *service.Service
}

func (s *TransactionServiceTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	s.SetupTransactionService()
	s.SetupCategoryService()

	gormDB := s.GormDB()

	mencache := trans_cache.NewMencache(s.GetCacheStore())
	repos := repository.NewRepositories(
		gormDB,
		pbcashier.NewCashierQueryServiceClient(s.Conns["cashier"]),
		pbmerchant.NewMerchantQueryServiceClient(s.Conns["merchant"]),
		pborder.NewOrderQueryServiceClient(s.Conns["order"]),
		pborderitem.NewOrderItemQueryServiceClient(s.Conns["order-item"]),
		pborderitem.NewOrderItemCommandServiceClient(s.Conns["order-item"]),
	)

	s.svc = service.NewService(&service.Deps{
		Kafka:         nil,
		Mencache:      mencache,
		Repositories:  repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
}

func (s *TransactionServiceTestSuite) TearDownSuite() {
	s.BaseTestSuite.TearDownSuite()
}

func (s *TransactionServiceTestSuite) TestTransactionLifecycle() {
	ctx := context.Background()

	userID := s.SeedUser(ctx)
	merchantID := s.SeedMerchant(ctx, userID)
	categoryID := s.SeedCategory(ctx)
	productID := s.SeedProduct(ctx, merchantID, categoryID)
	orderID := s.SeedOrder(ctx, userID, merchantID, productID)
	s.SeedOrderItem(ctx, orderID, productID)

	var cashierID int
	err := s.GormDB().Raw(`SELECT cashier_id FROM cashiers WHERE user_id = $1 AND merchant_id = $2 AND deleted_at IS NULL LIMIT 1`,
		userID, merchantID,
	).Scan(&cashierID).Error

	req := &requests.CreateTransactionRequest{
		OrderID:       orderID,
		CashierID:     cashierID,
		MerchantID:    merchantID,
		PaymentMethod: "Transfer Bank",
		Amount:        1000000,
	}
	created, err := s.svc.TransactionCommand.CreateTransaction(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(created)
	transactionID := int(created.TransactionID)

	found, err := s.svc.TransactionQuery.FindById(ctx, transactionID)
	s.Require().NoError(err)
	s.Require().NotNil(found.PaymentStatus)
	s.Equal("success", *found.PaymentStatus)

	newPaymentMethod := "GOPAY"
	updateReq := &requests.UpdateTransactionRequest{
		TransactionID: &transactionID,
		OrderID:       orderID,
		CashierID:     cashierID,
		PaymentMethod: newPaymentMethod,
		Amount:        req.Amount,
	}
	updated, err := s.svc.TransactionCommand.UpdateTransaction(ctx, updateReq)
	s.Require().NoError(err)
	s.Equal(newPaymentMethod, updated.PaymentMethod)

	_, total, err := s.svc.TransactionQuery.FindAllTransactions(ctx, &requests.FindAllTransaction{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*total, 1)

	_, err = s.svc.TransactionCommand.TrashedTransaction(ctx, transactionID)
	s.Require().NoError(err)

	_, totalTrashed, err := s.svc.TransactionQuery.FindByTrashed(ctx, &requests.FindAllTransaction{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*totalTrashed, 1)

	active, _, err := s.svc.TransactionQuery.FindByActive(ctx, &requests.FindAllTransaction{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	for _, tx := range active {
		s.NotEqual(transactionID, int(tx.TransactionID))
	}

	_, err = s.svc.TransactionCommand.RestoreTransaction(ctx, transactionID)
	s.Require().NoError(err)

	_, err = s.svc.TransactionCommand.TrashedTransaction(ctx, transactionID)
	s.Require().NoError(err)
	success, err := s.svc.TransactionCommand.DeleteTransactionPermanently(ctx, transactionID)
	s.Require().NoError(err)
	s.True(success)

	o1 := s.SeedOrder(ctx, userID, merchantID, productID)
	s.SeedOrderItem(ctx, o1, productID)

	o2 := s.SeedOrder(ctx, userID, merchantID, productID)
	s.SeedOrderItem(ctx, o2, productID)

	s.svc.TransactionCommand.TrashedTransaction(ctx, o1)
	s.svc.TransactionCommand.TrashedTransaction(ctx, o2)

	resRestoreAll, err := s.svc.TransactionCommand.RestoreAllTransactions(ctx)
	s.Require().NoError(err)
	s.True(resRestoreAll)

	s.svc.TransactionCommand.TrashedTransaction(ctx, o1)
	s.svc.TransactionCommand.TrashedTransaction(ctx, o2)

	resDeleteAll, err := s.svc.TransactionCommand.DeleteAllTransactionPermanent(ctx)
	s.Require().NoError(err)
	s.True(resDeleteAll)
}

func TestTransactionServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionServiceTestSuite))
}
