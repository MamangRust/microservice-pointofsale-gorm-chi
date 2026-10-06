package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-apigateway/apierror"
	gateway_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	"github.com/MamangRust/microservice-point-of-sale-apigateway/httpx"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pbstats "github.com/MamangRust/microservice-point-of-sale-pb/stats"
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	"github.com/go-chi/chi/v5"
)

// statsHandleApi exposes the stats-reader (ClickHouse) gRPC contracts as REST
// endpoints (F4 §7.4). All cache keys use the "apigw:" namespace so the gateway
// can never overwrite domain-service or stats-reader cache keys (§8.1).
type statsHandleApi struct {
	orderClient       pbstats.OrderStatsServiceClient
	cashierClient     pbstats.CashierStatsServiceClient
	categoryClient    pbstats.CategoryStatsServiceClient
	transactionClient pbstats.TransactionStatsStatusServiceClient
	logger            logger.LoggerInterface
	mapping           response_api.StatsResponseMapper
	apiHandler        apierror.ApiHandler
	cache             *gateway_cache.GatewayCache
}

func NewHandlerStats(
	router chi.Router,
	orderClient pbstats.OrderStatsServiceClient,
	cashierClient pbstats.CashierStatsServiceClient,
	categoryClient pbstats.CategoryStatsServiceClient,
	transactionClient pbstats.TransactionStatsStatusServiceClient,
	logger logger.LoggerInterface,
	mapping response_api.StatsResponseMapper,
	apiHandler apierror.ApiHandler,
	cache *gateway_cache.GatewayCache,
) *statsHandleApi {
	statsHandler := &statsHandleApi{
		orderClient:       orderClient,
		cashierClient:     cashierClient,
		categoryClient:    categoryClient,
		transactionClient: transactionClient,
		logger:            logger,
		mapping:           mapping,
		apiHandler:        apiHandler,
		cache:             cache,
	}

	router.Route("/api/order-stats", func(orderStats chi.Router) {
		orderStats.Get("/monthly-total-revenue", apiHandler.Handle("get-order-stats-monthly-total-revenue", statsHandler.FindMonthlyTotalRevenue))
		orderStats.Get("/yearly-total-revenue", apiHandler.Handle("get-order-stats-yearly-total-revenue", statsHandler.FindYearlyTotalRevenue))
		orderStats.Get("/monthly-revenue", apiHandler.Handle("get-order-stats-monthly-revenue", statsHandler.FindMonthlyRevenue))
		orderStats.Get("/yearly-revenue", apiHandler.Handle("get-order-stats-yearly-revenue", statsHandler.FindYearlyRevenue))
	})

	router.Route("/api/cashier-stats", func(cashierStats chi.Router) {
		cashierStats.Get("/monthly-total-sales", apiHandler.Handle("get-cashier-stats-monthly-total-sales", statsHandler.FindCashierMonthlyTotalSales))
		cashierStats.Get("/yearly-total-sales", apiHandler.Handle("get-cashier-stats-yearly-total-sales", statsHandler.FindCashierYearlyTotalSales))
		cashierStats.Get("/month-sales", apiHandler.Handle("get-cashier-stats-month-sales", statsHandler.FindCashierMonthSales))
		cashierStats.Get("/year-sales", apiHandler.Handle("get-cashier-stats-year-sales", statsHandler.FindCashierYearSales))
	})

	router.Route("/api/category-stats", func(categoryStats chi.Router) {
		categoryStats.Get("/monthly-total-price", apiHandler.Handle("get-category-stats-monthly-total-price", statsHandler.FindCategoryMonthlyTotalPrice))
		categoryStats.Get("/yearly-total-price", apiHandler.Handle("get-category-stats-yearly-total-price", statsHandler.FindCategoryYearlyTotalPrice))
		categoryStats.Get("/month-price", apiHandler.Handle("get-category-stats-month-price", statsHandler.FindCategoryMonthPrice))
		categoryStats.Get("/year-price", apiHandler.Handle("get-category-stats-year-price", statsHandler.FindCategoryYearPrice))
	})

	router.Route("/api/transaction-stats", func(transactionStats chi.Router) {
		transactionStats.Get("/monthly-success", apiHandler.Handle("get-transaction-stats-monthly-success", statsHandler.FindTransactionMonthlySuccess))
		transactionStats.Get("/yearly-success", apiHandler.Handle("get-transaction-stats-yearly-success", statsHandler.FindTransactionYearlySuccess))
		transactionStats.Get("/monthly-failed", apiHandler.Handle("get-transaction-stats-monthly-failed", statsHandler.FindTransactionMonthlyFailed))
		transactionStats.Get("/yearly-failed", apiHandler.Handle("get-transaction-stats-yearly-failed", statsHandler.FindTransactionYearlyFailed))
	})

	return statsHandler
}

func (h *statsHandleApi) parseYearMonth(r *http.Request) (int, int, error) {
	year, err := parseQueryIntWithValidation(r, "year", 1, 9999)
	if err != nil {
		return 0, 0, errors.NewBadRequestError("invalid year")
	}
	month, err := parseQueryIntWithValidation(r, "month", 1, 12)
	if err != nil {
		return 0, 0, errors.NewBadRequestError("invalid month")
	}
	return year, month, nil
}

func (h *statsHandleApi) parseYear(r *http.Request) (int, error) {
	year, err := parseQueryIntWithValidation(r, "year", 1, 9999)
	if err != nil {
		return 0, errors.NewBadRequestError("invalid year")
	}
	return year, nil
}

// @Summary Monthly total revenue (order)
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponseOrderMonthlyTotalRevenue
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/monthly-total-revenue [get]
// @Security BearerAuth
func (h *statsHandleApi) FindMonthlyTotalRevenue(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	year, month, err := h.parseYearMonth(r)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:order-stats:monthly-total-revenue:%d:%d", year, month)
	if cached, found := gateway_cache.Get[response.ApiResponseOrderMonthlyTotalRevenue](ctx, h.cache, cacheKey); found && cached != nil {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	res, err := h.orderClient.FindMonthlyTotalRevenue(ctx, &pborder.FindYearMonthTotalRevenue{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseOrderMonthlyTotalRevenue(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return httpx.JSON(w, http.StatusOK, so)
}

// @Summary Yearly total revenue (order)
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseOrderYearlyTotalRevenue
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/yearly-total-revenue [get]
// @Security BearerAuth
func (h *statsHandleApi) FindYearlyTotalRevenue(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	year, err := h.parseYear(r)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:order-stats:yearly-total-revenue:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseOrderYearlyTotalRevenue](ctx, h.cache, cacheKey); found && cached != nil {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	res, err := h.orderClient.FindYearlyTotalRevenue(ctx, &pborder.FindYearTotalRevenue{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseOrderYearlyTotalRevenue(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return httpx.JSON(w, http.StatusOK, so)
}

// @Summary Monthly revenue (order)
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseOrderMonthly
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/monthly-revenue [get]
// @Security BearerAuth
func (h *statsHandleApi) FindMonthlyRevenue(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	year, err := h.parseYear(r)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:order-stats:monthly-revenue:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseOrderMonthly](ctx, h.cache, cacheKey); found && cached != nil {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	res, err := h.orderClient.FindMonthlyRevenue(ctx, &pborder.FindYearOrder{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseOrderMonthly(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return httpx.JSON(w, http.StatusOK, so)
}

// @Summary Yearly revenue (order)
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseOrderYearly
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/yearly-revenue [get]
// @Security BearerAuth
func (h *statsHandleApi) FindYearlyRevenue(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	year, err := h.parseYear(r)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:order-stats:yearly-revenue:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseOrderYearly](ctx, h.cache, cacheKey); found && cached != nil {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	res, err := h.orderClient.FindYearlyRevenue(ctx, &pborder.FindYearOrder{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseOrderYearly(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return httpx.JSON(w, http.StatusOK, so)
}

// @Summary Monthly total sales (cashier)
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponseCashierMonthlyTotalSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/monthly-total-sales [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCashierMonthlyTotalSales(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	year, month, err := h.parseYearMonth(r)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:cashier-stats:monthly-total-sales:%d:%d", year, month)
	if cached, found := gateway_cache.Get[response.ApiResponseCashierMonthlyTotalSales](ctx, h.cache, cacheKey); found && cached != nil {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	res, err := h.cashierClient.FindMonthlyTotalSales(ctx, &pbcashier.FindYearMonthTotalSales{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierMonthlyTotalSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return httpx.JSON(w, http.StatusOK, so)
}

// @Summary Yearly total sales (cashier)
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCashierYearlyTotalSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/yearly-total-sales [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCashierYearlyTotalSales(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	year, err := h.parseYear(r)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:cashier-stats:yearly-total-sales:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseCashierYearlyTotalSales](ctx, h.cache, cacheKey); found && cached != nil {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	res, err := h.cashierClient.FindYearlyTotalSales(ctx, &pbcashier.FindYearTotalSales{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierYearlyTotalSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return httpx.JSON(w, http.StatusOK, so)
}

// @Summary Month sales (cashier)
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCashierMonthSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/month-sales [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCashierMonthSales(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	year, err := h.parseYear(r)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:cashier-stats:month-sales:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseCashierMonthSales](ctx, h.cache, cacheKey); found && cached != nil {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	res, err := h.cashierClient.FindMonthSales(ctx, &pbcashier.FindYearCashier{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierMonthSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return httpx.JSON(w, http.StatusOK, so)
}

// @Summary Year sales (cashier)
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCashierYearSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/year-sales [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCashierYearSales(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	year, err := h.parseYear(r)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:cashier-stats:year-sales:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseCashierYearSales](ctx, h.cache, cacheKey); found && cached != nil {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	res, err := h.cashierClient.FindYearSales(ctx, &pbcashier.FindYearCashier{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierYearSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return httpx.JSON(w, http.StatusOK, so)
}

// @Summary Monthly total price (category)
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponseCategoryMonthlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/monthly-total-price [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCategoryMonthlyTotalPrice(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	year, month, err := h.parseYearMonth(r)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:category-stats:monthly-total-price:%d:%d", year, month)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryMonthlyTotalPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	res, err := h.categoryClient.FindMonthlyTotalPrices(ctx, &pbcategory.FindYearMonthTotalPrices{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryMonthlyTotalPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return httpx.JSON(w, http.StatusOK, so)
}

// @Summary Yearly total price (category)
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCategoryYearlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/yearly-total-price [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCategoryYearlyTotalPrice(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	year, err := h.parseYear(r)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:category-stats:yearly-total-price:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryYearlyTotalPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	res, err := h.categoryClient.FindYearlyTotalPrices(ctx, &pbcategory.FindYearTotalPrices{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryYearlyTotalPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return httpx.JSON(w, http.StatusOK, so)
}

// @Summary Month price (category)
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCategoryMonthPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/month-price [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCategoryMonthPrice(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	year, err := h.parseYear(r)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:category-stats:month-price:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryMonthPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	res, err := h.categoryClient.FindMonthPrice(ctx, &pbcategory.FindYearCategory{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryMonthPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return httpx.JSON(w, http.StatusOK, so)
}

// @Summary Year price (category)
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCategoryYearPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/year-price [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCategoryYearPrice(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	year, err := h.parseYear(r)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:category-stats:year-price:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryYearPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	res, err := h.categoryClient.FindYearPrice(ctx, &pbcategory.FindYearCategory{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryYearPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return httpx.JSON(w, http.StatusOK, so)
}

// @Summary Monthly success (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponsesTransactionMonthSuccess
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/monthly-success [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionMonthlySuccess(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	year, month, err := h.parseYearMonth(r)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:monthly-success:%d:%d", year, month)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionMonthSuccess](ctx, h.cache, cacheKey); found && cached != nil {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	res, err := h.transactionClient.FindMonthStatusSuccess(ctx, &pbtransaction.FindMonthlyTransactionStatus{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseTransactionMonthSuccess(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return httpx.JSON(w, http.StatusOK, so)
}

// @Summary Yearly success (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponsesTransactionYearSuccess
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/yearly-success [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionYearlySuccess(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	year, err := h.parseYear(r)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:yearly-success:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionYearSuccess](ctx, h.cache, cacheKey); found && cached != nil {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	res, err := h.transactionClient.FindYearStatusSuccess(ctx, &pbtransaction.FindYearlyTransactionStatus{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseTransactionYearSuccess(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return httpx.JSON(w, http.StatusOK, so)
}

// @Summary Monthly failed (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponsesTransactionMonthFailed
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/monthly-failed [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionMonthlyFailed(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	year, month, err := h.parseYearMonth(r)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:monthly-failed:%d:%d", year, month)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionMonthFailed](ctx, h.cache, cacheKey); found && cached != nil {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	res, err := h.transactionClient.FindMonthStatusFailed(ctx, &pbtransaction.FindMonthlyTransactionStatus{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseTransactionMonthFailed(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return httpx.JSON(w, http.StatusOK, so)
}

// @Summary Yearly failed (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponsesTransactionYearFailed
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/yearly-failed [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionYearlyFailed(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	year, err := h.parseYear(r)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:yearly-failed:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionYearFailed](ctx, h.cache, cacheKey); found && cached != nil {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	res, err := h.transactionClient.FindYearStatusFailed(ctx, &pbtransaction.FindYearlyTransactionStatus{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseTransactionYearFailed(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return httpx.JSON(w, http.StatusOK, so)
}
