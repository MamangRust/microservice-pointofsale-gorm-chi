package service

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/kafka"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	mencache "github.com/MamangRust/microservice-point-of-sale-transacton/cache"
	"github.com/MamangRust/microservice-point-of-sale-transacton/repository"
	"gorm.io/gorm"
)

type Service struct {
	TransactionQuery   TransactionQueryService
	TransactionCommand TransactionCommandService
}

type Deps struct {
	Ctx           context.Context
	Kafka         *kafka.Kafka
	Mencache      mencache.Mencache
	Repositories  *repository.Repositories
	GormDB        *gorm.DB
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

func NewService(deps *Deps) *Service {
	var kafkaPublisher EmailEventPublisher
	if deps.Kafka != nil {
		kafkaPublisher = deps.Kafka
	}

	return &Service{
		TransactionQuery:   NewTransactionQueryService(deps.Mencache, deps.Repositories.TransactionQueryRepository, deps.Logger, deps.Observability),
		TransactionCommand: NewTransactionCommandService(kafkaPublisher, deps.Mencache, deps.Repositories.CashierQuery, deps.Repositories.MerchantQuery, deps.Repositories.TransactionQueryRepository, deps.Repositories.TransactionCommandRepository, deps.Repositories.OrderQuery, deps.Repositories.OrderItemQuery, deps.Logger, deps.Observability),
	}
}
