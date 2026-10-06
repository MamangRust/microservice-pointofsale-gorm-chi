package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-point-of-sale-apigateway/apierror"
	auth_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/auth"
	gateway_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	"github.com/MamangRust/microservice-point-of-sale-pkg/auth"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-pkg/upload_image"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"

	pbauth "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pbmerchantdoc "github.com/MamangRust/microservice-point-of-sale-pb/merchant_document"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbstats "github.com/MamangRust/microservice-point-of-sale-pb/stats"
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
)

type ServiceConnections struct {
	Auth        *grpc.ClientConn
	Role        *grpc.ClientConn
	User        *grpc.ClientConn
	Cashier     *grpc.ClientConn
	Category    *grpc.ClientConn
	Merchant    *grpc.ClientConn
	OrderItem   *grpc.ClientConn
	Order       *grpc.ClientConn
	Product     *grpc.ClientConn
	Transaction *grpc.ClientConn
	StatsReader *grpc.ClientConn
}

type Deps struct {
	Token              auth.TokenManager
	Router             chi.Router
	Logger             logger.LoggerInterface
	Mapping            *response_api.ResponseApiMapper
	ImageUpload        upload_image.ImageUploads
	ServiceConnections *ServiceConnections
	ApiHandler         apierror.ApiHandler
	AuthCache          auth_cache.AuthMencache
	GatewayCache       *gateway_cache.GatewayCache
}

func NewHandler(deps *Deps) {
	clientAuth := pbauth.NewAuthServiceClient(deps.ServiceConnections.Auth)
	clientRole := pbrole.NewRoleQueryServiceClient(deps.ServiceConnections.Role)
	clientRoleCommand := pbrole.NewRoleCommandServiceClient(deps.ServiceConnections.Role)
	clientUser := pbuser.NewUserQueryServiceClient(deps.ServiceConnections.User)
	clientUserCommand := pbuser.NewUserCommandServiceClient(deps.ServiceConnections.User)
	clientUserRole := pbuserrole.NewUserRoleServiceClient(deps.ServiceConnections.Role)
	clientCategory := pbcategory.NewCategoryQueryServiceClient(deps.ServiceConnections.Category)
	clientCategoryCommand := pbcategory.NewCategoryCommandServiceClient(deps.ServiceConnections.Category)
	clientCashier := pbcashier.NewCashierQueryServiceClient(deps.ServiceConnections.Cashier)
	clientCashierCommand := pbcashier.NewCashierCommandServiceClient(deps.ServiceConnections.Cashier)
	clientMerchant := pbmerchant.NewMerchantQueryServiceClient(deps.ServiceConnections.Merchant)
	clientMerchantCommand := pbmerchant.NewMerchantCommandServiceClient(deps.ServiceConnections.Merchant)
	clientMerchantDoc := pbmerchantdoc.NewMerchantDocumentServiceClient(deps.ServiceConnections.Merchant)
	clientOrderItem := pborderitem.NewOrderItemQueryServiceClient(deps.ServiceConnections.OrderItem)
	clientOrder := pborder.NewOrderQueryServiceClient(deps.ServiceConnections.Order)
	clientOrderCommand := pborder.NewOrderCommandServiceClient(deps.ServiceConnections.Order)
	clientProduct := pbproduct.NewProductQueryServiceClient(deps.ServiceConnections.Product)
	clientProductCommand := pbproduct.NewProductCommandServiceClient(deps.ServiceConnections.Product)
	clientTransaction := pbtransaction.NewTransactionQueryServiceClient(deps.ServiceConnections.Transaction)
	clientTransactionCommand := pbtransaction.NewTransactionCommandServiceClient(deps.ServiceConnections.Transaction)

	NewHandlerAuth(deps.Router, clientAuth, deps.Logger, deps.Mapping.AuthResponseMapper, deps.ApiHandler, deps.AuthCache)
	NewHandlerRole(deps.Router, clientRole, clientRoleCommand, clientUserRole, deps.Logger, deps.Mapping.RoleResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerUser(deps.Router, clientUser, clientUserCommand, deps.Logger, deps.Mapping.UserResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerCategory(deps.Router, clientCategory, clientCategoryCommand, deps.Logger, deps.Mapping.CategoryResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerCashier(deps.Router, clientCashier, clientCashierCommand, deps.Logger, deps.Mapping.CashierResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerMerchant(deps.Router, clientMerchant, clientMerchantCommand, deps.Logger, deps.Mapping.MerchantResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerMerchantDocument(deps.Router, clientMerchantDoc, deps.Logger, deps.Mapping.MerchantDocumentProMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerOrderItem(deps.Router, clientOrderItem, deps.Logger, deps.Mapping.OrderItemResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerOrder(deps.Router, clientOrder, clientOrderCommand, deps.Logger, deps.Mapping.OrderResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerProduct(deps.Router, clientProduct, clientProductCommand, deps.Logger, deps.Mapping.ProductResponseMapper, deps.ImageUpload, deps.ApiHandler, deps.GatewayCache)
	NewHandlerTransaction(deps.Router, clientTransaction, clientTransactionCommand, deps.Logger, deps.Mapping.TransactionResponseMapper, deps.ApiHandler, deps.GatewayCache)

	clientOrderStats := pbstats.NewOrderStatsServiceClient(deps.ServiceConnections.StatsReader)
	clientCashierStats := pbstats.NewCashierStatsServiceClient(deps.ServiceConnections.StatsReader)
	clientCategoryStats := pbstats.NewCategoryStatsServiceClient(deps.ServiceConnections.StatsReader)
	clientTransactionStats := pbstats.NewTransactionStatsStatusServiceClient(deps.ServiceConnections.StatsReader)

	NewHandlerStats(deps.Router, clientOrderStats, clientCashierStats, clientCategoryStats, clientTransactionStats, deps.Logger, deps.Mapping.StatsResponseMapper, deps.ApiHandler, deps.GatewayCache)
}

func parseQueryInt(r *http.Request, key string, defaultValue int) int {
	val, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil || val <= 0 {
		return defaultValue
	}
	return val
}

func parseQueryIntWithValidation(r *http.Request, key string, min, max int) (int, error) {
	valStr := r.URL.Query().Get(key)
	val, err := strconv.Atoi(valStr)
	if err != nil || val < min || val > max {
		return 0, fmt.Errorf("invalid %s: %s", key, valStr)
	}
	return val, nil
}
