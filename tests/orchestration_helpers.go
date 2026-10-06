package tests

import (
	"bytes"
	"context"
	"mime/multipart"

	pbauth "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pbmerchantdoc "github.com/MamangRust/microservice-point-of-sale-pb/merchant_document"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/auth"
	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Role
	role_cache "github.com/MamangRust/microservice-point-of-sale-role/cache"
	role_handler "github.com/MamangRust/microservice-point-of-sale-role/handler"
	role_repo "github.com/MamangRust/microservice-point-of-sale-role/repository"
	role_service "github.com/MamangRust/microservice-point-of-sale-role/service"

	// User
	user_cache "github.com/MamangRust/microservice-point-of-sale-user/cache"
	user_handler "github.com/MamangRust/microservice-point-of-sale-user/handler"
	user_repo "github.com/MamangRust/microservice-point-of-sale-user/repository"
	user_service "github.com/MamangRust/microservice-point-of-sale-user/service"

	// Auth
	auth_cache "github.com/MamangRust/microservice-point-of-sale-auth/cache"
	auth_handler "github.com/MamangRust/microservice-point-of-sale-auth/handler"
	auth_repo "github.com/MamangRust/microservice-point-of-sale-auth/repository"
	auth_service "github.com/MamangRust/microservice-point-of-sale-auth/service"

	// Category
	category_cache "github.com/MamangRust/microservice-point-of-sale-category/cache"
	category_handler "github.com/MamangRust/microservice-point-of-sale-category/handler"
	category_repo "github.com/MamangRust/microservice-point-of-sale-category/repository"
	category_service "github.com/MamangRust/microservice-point-of-sale-category/service"

	// Product
	product_cache "github.com/MamangRust/microservice-point-of-sale-product/cache"
	product_handler "github.com/MamangRust/microservice-point-of-sale-product/handler"
	product_repo "github.com/MamangRust/microservice-point-of-sale-product/repository"
	product_service "github.com/MamangRust/microservice-point-of-sale-product/service"

	// Merchant
	merchant_cache "github.com/MamangRust/microservice-point-of-sale-merchant/cache"
	merchant_handler "github.com/MamangRust/microservice-point-of-sale-merchant/handler"
	merchant_repo "github.com/MamangRust/microservice-point-of-sale-merchant/repository"
	merchant_service "github.com/MamangRust/microservice-point-of-sale-merchant/service"

	// Order
	order_cache "github.com/MamangRust/microservice-point-of-sale-order/cache"
	order_handler "github.com/MamangRust/microservice-point-of-sale-order/handler"
	order_repo "github.com/MamangRust/microservice-point-of-sale-order/repository"
	order_service "github.com/MamangRust/microservice-point-of-sale-order/service"

	// Transaction
	transaction_cache "github.com/MamangRust/microservice-point-of-sale-transacton/cache"
	transaction_handler "github.com/MamangRust/microservice-point-of-sale-transacton/handler"
	transaction_repo "github.com/MamangRust/microservice-point-of-sale-transacton/repository"
	transaction_service "github.com/MamangRust/microservice-point-of-sale-transacton/service"

	// Order Item
	order_item_cache "github.com/MamangRust/microservice-point-of-sale-order-item/cache"
	order_item_handler "github.com/MamangRust/microservice-point-of-sale-order-item/handler"
	order_item_repo "github.com/MamangRust/microservice-point-of-sale-order-item/repository"
	order_item_service "github.com/MamangRust/microservice-point-of-sale-order-item/service"

	// Cashier
	mencache "github.com/MamangRust/microservice-point-of-sale-cashier/cache"
	cashier_handler "github.com/MamangRust/microservice-point-of-sale-cashier/handler"
	cashier_repo "github.com/MamangRust/microservice-point-of-sale-cashier/repository"
	cashier_service "github.com/MamangRust/microservice-point-of-sale-cashier/service"
)

func (s *BaseTestSuite) SetupRoleService() {
	cacheStore := s.GetCacheStore()
	gormDB := s.GormDB()

	roleMencache := role_cache.NewMencache(cacheStore)
	roleRepos := role_repo.NewRepositories(gormDB)
	roleSvc := role_service.NewService(&role_service.Deps{
		Repositories:  roleRepos,
		Logger:        s.Log,
		Mencache:      roleMencache,
		Observability: s.Obs,
	})
	roleGapi := role_handler.NewHandler(roleSvc)
	server := grpc.NewServer()
	pbrole.RegisterRoleQueryServiceServer(server, roleGapi)
	pbrole.RegisterRoleCommandServiceServer(server, roleGapi)
	pbuserrole.RegisterUserRoleServiceServer(server, roleGapi)
	addr, err := RunGRPCServer(server)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.Conns["role"] = conn
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupUserService() {
	if _, ok := s.Conns["role"]; !ok {
		s.SetupRoleService()
	}

	cacheStore := s.GetCacheStore()
	gormDB := s.GormDB()
	hasher := hash.NewHashingPassword()

	roleClient := pbrole.NewRoleQueryServiceClient(s.Conns["role"])

	userMencache := user_cache.NewMencache(cacheStore)
	userRepos := user_repo.NewRepositories(gormDB, roleClient, pbuserrole.NewUserRoleServiceClient(s.Conns["role"]))
	userSvc := user_service.NewService(&user_service.Deps{
		Repositories:  userRepos,
		Logger:        s.Log,
		Hash:          hasher,
		Mencache:      userMencache,
		Observability: s.Obs,
	})
	userGapi := user_handler.NewHandler(userSvc)
	server := grpc.NewServer()
	pbuser.RegisterUserQueryServiceServer(server, userGapi)
	pbuser.RegisterUserCommandServiceServer(server, userGapi)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.Conns["user"] = conn
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupAuthService() {
	if _, ok := s.Conns["role"]; !ok {
		s.SetupRoleService()
	}
	if _, ok := s.Conns["user"]; !ok {
		s.SetupUserService()
	}

	cacheStore := s.GetCacheStore()
	gormDB := s.GormDB()
	hasher := hash.NewHashingPassword()
	tokenManager, _ := auth.NewManager("mysecret")

	roleClient := pbrole.NewRoleQueryServiceClient(s.Conns["role"])
	userRoleClient := pbuserrole.NewUserRoleServiceClient(s.Conns["role"])
	userQueryClient := pbuser.NewUserQueryServiceClient(s.Conns["user"])
	userCommandClient := pbuser.NewUserCommandServiceClient(s.Conns["user"])

	authRepos := auth_repo.NewRepositories(gormDB, userQueryClient, userCommandClient, roleClient, userRoleClient)
	authMencache := auth_cache.NewMencache(cacheStore)
	authSvc := auth_service.NewService(&auth_service.Deps{
		Repositories:  authRepos,
		Logger:        s.Log,
		Mencache:      authMencache,
		Token:         tokenManager,
		Hash:          hasher,
		Kafka:         nil,
		Observability: s.Obs,
	})
	authGapi := auth_handler.NewAuthHandleGrpc(authSvc, s.Log)
	server := grpc.NewServer()
	pbauth.RegisterAuthServiceServer(server, authGapi)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.Conns["auth"] = conn
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupCategoryService() {
	cacheStore := s.GetCacheStore()
	gormDB := s.GormDB()

	catMencache := category_cache.NewMencache(cacheStore)
	catRepos := category_repo.NewRepositories(gormDB)
	catSvc := category_service.NewService(&category_service.Deps{
		Mencache:      catMencache,
		Repositories:  catRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	catGapi := category_handler.NewHandler(catSvc)
	server := grpc.NewServer()
	pbcategory.RegisterCategoryQueryServiceServer(server, catGapi)
	pbcategory.RegisterCategoryCommandServiceServer(server, catGapi)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["category"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupProductService() {
	if _, ok := s.Conns["category"]; !ok {
		s.SetupCategoryService()
	}
	if _, ok := s.Conns["merchant"]; !ok {
		s.SetupMerchantService()
	}

	cacheStore := s.GetCacheStore()
	gormDB := s.GormDB()

	categoryClient := pbcategory.NewCategoryQueryServiceClient(s.Conns["category"])
	merchantClient := pbmerchant.NewMerchantQueryServiceClient(s.Conns["merchant"])

	prodMencache := product_cache.NewMencache(cacheStore)
	prodRepos := product_repo.NewRepositories(gormDB, categoryClient, merchantClient)
	prodSvc := product_service.NewService(&product_service.Deps{
		Mencache:      prodMencache,
		Repositories:  prodRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	prodGapi := product_handler.NewHandler(prodSvc)
	server := grpc.NewServer()
	pbproduct.RegisterProductQueryServiceServer(server, prodGapi)
	pbproduct.RegisterProductCommandServiceServer(server, prodGapi)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["product"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupMerchantService() {
	if _, ok := s.Conns["user"]; !ok {
		s.SetupUserService()
	}

	cacheStore := s.GetCacheStore()
	gormDB := s.GormDB()

	merchantMencache := merchant_cache.NewMencache(cacheStore)
	userQueryClient := pbuser.NewUserQueryServiceClient(s.Conns["user"])
	userCommandClient := pbuser.NewUserCommandServiceClient(s.Conns["user"])
	merchantRepos := merchant_repo.NewRepositories(gormDB, userQueryClient, userCommandClient)
	merchantSvc := merchant_service.NewService(&merchant_service.Deps{
		Mencache:      merchantMencache,
		Repositories:  merchantRepos,
		Logger:        s.Log,
		Observability: s.Obs,
		Kafka:         nil,
	})
	merchantHandler, merchantDocHandler := merchant_handler.NewHandler(merchantSvc)
	server := grpc.NewServer()
	pbmerchant.RegisterMerchantQueryServiceServer(server, merchantHandler)
	pbmerchant.RegisterMerchantCommandServiceServer(server, merchantHandler)
	pbmerchantdoc.RegisterMerchantDocumentServiceServer(server, merchantDocHandler)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["merchant"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupOrderService() {
	if _, ok := s.Conns["cashier"]; !ok {
		s.SetupCashierService()
	}
	if _, ok := s.Conns["merchant"]; !ok {
		s.SetupMerchantService()
	}
	if _, ok := s.Conns["product"]; !ok {
		s.SetupProductService()
	}
	if _, ok := s.Conns["order-item"]; !ok {
		s.SetupOrderItemService()
	}

	cacheStore := s.GetCacheStore()
	gormDB := s.GormDB()

	cashierClient := pbcashier.NewCashierQueryServiceClient(s.Conns["cashier"])
	merchantClient := pbmerchant.NewMerchantQueryServiceClient(s.Conns["merchant"])
	productQueryClient := pbproduct.NewProductQueryServiceClient(s.Conns["product"])
	productCommandClient := pbproduct.NewProductCommandServiceClient(s.Conns["product"])
	orderItemQueryClient := pborderitem.NewOrderItemQueryServiceClient(s.Conns["order-item"])
	orderItemCommandClient := pborderitem.NewOrderItemCommandServiceClient(s.Conns["order-item"])

	orderMencache := order_cache.NewMencache(cacheStore)
	orderRepos := order_repo.NewRepositories(gormDB, cashierClient, merchantClient, productQueryClient, productCommandClient, orderItemQueryClient, orderItemCommandClient)
	orderSvc := order_service.NewService(&order_service.Deps{
		Mencache:      orderMencache,
		Repositories:  orderRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	orderGapi := order_handler.NewHandler(orderSvc)
	server := grpc.NewServer()
	pborder.RegisterOrderQueryServiceServer(server, orderGapi)
	pborder.RegisterOrderCommandServiceServer(server, orderGapi)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["order"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupTransactionService() {
	if _, ok := s.Conns["cashier"]; !ok {
		s.SetupCashierService()
	}
	if _, ok := s.Conns["merchant"]; !ok {
		s.SetupMerchantService()
	}
	if _, ok := s.Conns["order"]; !ok {
		s.SetupOrderService()
	}
	if _, ok := s.Conns["order-item"]; !ok {
		s.SetupOrderItemService()
	}

	cacheStore := s.GetCacheStore()
	gormDB := s.GormDB()

	cashierClient := pbcashier.NewCashierQueryServiceClient(s.Conns["cashier"])
	merchantClient := pbmerchant.NewMerchantQueryServiceClient(s.Conns["merchant"])
	orderClient := pborder.NewOrderQueryServiceClient(s.Conns["order"])
	orderItemQueryClient := pborderitem.NewOrderItemQueryServiceClient(s.Conns["order-item"])
	orderItemCommandClient := pborderitem.NewOrderItemCommandServiceClient(s.Conns["order-item"])

	transactionMencache := transaction_cache.NewMencache(cacheStore)
	transactionRepos := transaction_repo.NewRepositories(gormDB, cashierClient, merchantClient, orderClient, orderItemQueryClient, orderItemCommandClient)
	transactionSvc := transaction_service.NewService(&transaction_service.Deps{
		Mencache:      transactionMencache,
		Repositories:  transactionRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	transactionGapi := transaction_handler.NewHandler(transactionSvc, s.Log)
	server := grpc.NewServer()
	pbtransaction.RegisterTransactionQueryServiceServer(server, transactionGapi)
	pbtransaction.RegisterTransactionCommandServiceServer(server, transactionGapi)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["transaction"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupOrderItemService() {
	cacheStore := s.GetCacheStore()
	gormDB := s.GormDB()

	itemMencache := order_item_cache.NewMencache(cacheStore)
	itemRepos := order_item_repo.NewRepositories(gormDB)
	itemSvc := order_item_service.NewService(&order_item_service.Deps{
		Mencache:      itemMencache,
		Repositories:  itemRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	itemGapi := order_item_handler.NewHandler(itemSvc, s.Log)
	server := grpc.NewServer()
	pborderitem.RegisterOrderItemQueryServiceServer(server, itemGapi)
	pborderitem.RegisterOrderItemCommandServiceServer(server, itemGapi)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["order-item"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupCashierService() {
	if _, ok := s.Conns["user"]; !ok {
		s.SetupUserService()
	}
	if _, ok := s.Conns["merchant"]; !ok {
		s.SetupMerchantService()
	}

	cacheStore := s.GetCacheStore()
	gormDB := s.GormDB()

	cashierMencache := mencache.NewMencache(cacheStore)
	userQueryClient := pbuser.NewUserQueryServiceClient(s.Conns["user"])
	userCommandClient := pbuser.NewUserCommandServiceClient(s.Conns["user"])
	merchantClient := pbmerchant.NewMerchantQueryServiceClient(s.Conns["merchant"])
	cashierRepos := cashier_repo.NewRepositories(gormDB, userQueryClient, userCommandClient, merchantClient)
	cashierSvc := cashier_service.NewService(&cashier_service.Deps{
		Ctx:           context.Background(),
		Mencache:      cashierMencache,
		Repositories:  cashierRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	cashierGapi := cashier_handler.NewHandler(cashierSvc)
	server := grpc.NewServer()
	pbcashier.RegisterCashierQueryServiceServer(server, cashierGapi)
	pbcashier.RegisterCashierCommandServiceServer(server, cashierGapi)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["cashier"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) dial(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	return conn
}

func (s *BaseTestSuite) GetCacheStore() *cache.CacheStore {
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	return cache.NewCacheStore(s.ts.RedisClient(), s.Log, cacheMetrics)
}

func (s *BaseTestSuite) BuildMultipartRequestBody(fields map[string]string, fieldName, fileName string) ([]byte, string) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range fields {
		fw, _ := w.CreateFormField(key)
		fw.Write([]byte(r))
	}
	fw, _ := w.CreateFormFile(fieldName, fileName)
	fw.Write([]byte("dummy image content"))
	w.Close()
	return b.Bytes(), w.FormDataContentType()
}
