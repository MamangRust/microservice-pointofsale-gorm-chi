package mencache

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
)

type orderItemCommandCache struct {
	store *cache.CacheStore
}

func NewOrderItemCommandCache(store *cache.CacheStore) OrderItemCommandCache {
	return &orderItemCommandCache{store: store}
}

func (o *orderItemCommandCache) DeleteCachedOrderItems(ctx context.Context, orderID int) {
	cache.DeleteFromCache(ctx, o.store, fmt.Sprintf(orderItemByIdCacheKey, orderID))
}

func (o *orderItemCommandCache) DeleteCachedOrderItemsAllCache(ctx context.Context) {
	o.store.InvalidateCache(ctx, "order_item:*")
}
