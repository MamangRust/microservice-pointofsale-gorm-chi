package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/exp/rand"
	"gorm.io/gorm"
)

type categorySeeder struct {
	db     *gorm.DB
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewCategorySeeder(db *gorm.DB, ctx context.Context, logger logger.LoggerInterface) *categorySeeder {
	return &categorySeeder{db: db, ctx: ctx, logger: logger}
}

func (r *categorySeeder) Seed() error {
	categoryNames := []string{
		"Electronics", "Clothing", "Groceries", "Toys", "Home & Kitchen",
		"Books", "Beauty & Health", "Sports & Outdoors", "Automotive", "Furniture",
	}
	categoryDescriptions := []string{
		"Best electronics products", "Latest fashion trends", "Fresh groceries",
		"Fun toys for kids", "Essentials for home & kitchen",
		"Books for all ages", "Beauty and health products",
		"Outdoor sports equipment", "Automotive accessories", "Stylish furniture",
	}

	now := time.Now()
	for i := 0; i < 10; i++ {
		name := categoryNames[i%len(categoryNames)]
		description := categoryDescriptions[i%len(categoryDescriptions)]
		slugCategory := fmt.Sprintf("%s-%d", name, i+1)

		err := r.db.WithContext(r.ctx).Exec(
			`INSERT INTO categories (name, description, slug_category, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
			name, description, slugCategory, now, now,
		).Error
		if err != nil {
			r.logger.Error("Failed to create category:", zap.Error(err))
			return err
		}
		r.logger.Debug("Category created:", zap.String("slug", slugCategory))
	}

	r.logger.Info("Category seeding completed successfully.", zap.Int("count", 10))
	return nil
}

// merchantRow is shared by the cashier/product/order/transaction seeders.
type merchantRow struct {
	MerchantID int
}

type cashierSeeder struct {
	merchantDB *gorm.DB
	identityDB *gorm.DB
	ctx        context.Context
	logger     logger.LoggerInterface
}

func NewCashierSeeder(merchantDB, identityDB *gorm.DB, ctx context.Context, logger logger.LoggerInterface) *cashierSeeder {
	return &cashierSeeder{merchantDB: merchantDB, identityDB: identityDB, ctx: ctx, logger: logger}
}

func (r *cashierSeeder) Seed() error {
	var merchants []merchantRow
	err := r.merchantDB.WithContext(r.ctx).Raw(`SELECT merchant_id FROM merchants WHERE deleted_at IS NULL LIMIT 20`).Scan(&merchants).Error
	if err != nil {
		r.logger.Error("Failed to fetch merchants:", zap.Error(err))
		return err
	}

	var users []userRow
	err = r.identityDB.WithContext(r.ctx).Raw(`SELECT user_id FROM users WHERE deleted_at IS NULL LIMIT 20`).Scan(&users).Error
	if err != nil {
		r.logger.Error("Failed to fetch users:", zap.Error(err))
		return err
	}

	if len(merchants) == 0 || len(users) == 0 {
		r.logger.Error("Merchants or Users not found. Seed operation aborted.")
		return fmt.Errorf("no merchants or users found")
	}

	now := time.Now()
	for i := 1; i <= 10; i++ {
		merchant := merchants[rand.Intn(len(merchants))]
		user := users[rand.Intn(len(users))]
		cashierName := fmt.Sprintf("Cashier %d", i)

		err := r.merchantDB.WithContext(r.ctx).Exec(
			`INSERT INTO cashiers (merchant_id, user_id, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
			merchant.MerchantID, user.UserID, cashierName, now, now,
		).Error
		if err != nil {
			r.logger.Error("Failed to create cashier:", zap.Error(err))
			return err
		}
	}

	r.logger.Info("Cashier seeding completed successfully.", zap.Int("count", 10))
	return nil
}

type merchantSeeder struct {
	merchantDB *gorm.DB
	identityDB *gorm.DB
	ctx        context.Context
	logger     logger.LoggerInterface
}

func NewMerchantSeeder(merchantDB, identityDB *gorm.DB, ctx context.Context, logger logger.LoggerInterface) *merchantSeeder {
	return &merchantSeeder{merchantDB: merchantDB, identityDB: identityDB, ctx: ctx, logger: logger}
}

func (r *merchantSeeder) Seed() error {
	var users []userRow
	err := r.identityDB.WithContext(r.ctx).Raw(`SELECT user_id FROM users WHERE deleted_at IS NULL LIMIT 20`).Scan(&users).Error
	if err != nil {
		r.logger.Error("Failed to fetch users:", zap.Error(err))
		return err
	}

	now := time.Now()
	for i := 1; i <= 10; i++ {
		userID := users[i%len(users)].UserID

		err = r.merchantDB.WithContext(r.ctx).Exec(
			`INSERT INTO merchants (user_id, name, description, address, contact_email, contact_phone, status, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			userID,
			fmt.Sprintf("Toko %d", i),
			fmt.Sprintf("Deskripsi untuk Toko %d", i),
			fmt.Sprintf("Jl. Toko %d", i),
			fmt.Sprintf("toko%d@example.com", i),
			fmt.Sprintf("0812345678%d", i),
			"active",
			now, now,
		).Error
		if err != nil {
			r.logger.Error("Failed to create merchant:", zap.Error(err))
			return err
		}
	}

	r.logger.Info("Merchant seeding completed successfully.", zap.Int("count", 10))
	return nil
}

type productSeeder struct {
	catalogDB  *gorm.DB
	merchantDB *gorm.DB
	ctx        context.Context
	logger     logger.LoggerInterface
}

func NewProductSeeder(catalogDB, merchantDB *gorm.DB, ctx context.Context, logger logger.LoggerInterface) *productSeeder {
	return &productSeeder{catalogDB: catalogDB, merchantDB: merchantDB, ctx: ctx, logger: logger}
}

type categoryRow struct {
	CategoryID int
}

func (r *productSeeder) Seed() error {
	var merchants []merchantRow
	err := r.merchantDB.WithContext(r.ctx).Raw(`SELECT merchant_id FROM merchants WHERE deleted_at IS NULL LIMIT 20`).Scan(&merchants).Error
	if err != nil {
		r.logger.Error("Failed to get merchants:", zap.Error(err))
		return err
	}

	var categories []categoryRow
	err = r.catalogDB.WithContext(r.ctx).Raw(`SELECT category_id FROM categories WHERE deleted_at IS NULL LIMIT 20`).Scan(&categories).Error
	if err != nil {
		r.logger.Error("Failed to get categories:", zap.Error(err))
		return err
	}

	if len(merchants) == 0 || len(categories) == 0 {
		r.logger.Error("No merchants or categories found, skipping seeding")
		return nil
	}

	productNames := []string{
		"Smartphone", "Laptop", "Wireless Earbuds", "Gaming Mouse", "Mechanical Keyboard",
		"Smartwatch", "Power Bank", "Bluetooth Speaker", "External Hard Drive", "Monitor",
	}
	brands := []string{"Samsung", "Apple", "Sony", "Logitech", "Razer", "Xiaomi", "HP", "Dell", "Acer", "Asus"}
	images := []string{
		"image1.jpg", "image2.jpg", "image3.jpg", "image4.jpg", "image5.jpg",
		"image6.jpg", "image7.jpg", "image8.jpg", "image9.jpg", "image10.jpg",
	}

	now := time.Now()
	for i := 0; i < 10; i++ {
		merchant := merchants[rand.Intn(len(merchants))]
		category := categories[rand.Intn(len(categories))]
		name := productNames[rand.Intn(len(productNames))]
		brand := brands[rand.Intn(len(brands))]
		price := int32(rand.Intn(5000000) + 50000)
		countInStock := int32(rand.Intn(100) + 1)
		weight := int32(rand.Intn(5000) + 100)
		slug := fmt.Sprintf("%s-%d", name, rand.Intn(1000))
		image := images[rand.Intn(len(images))]
		barcode := fmt.Sprintf("BC-%d", rand.Intn(9999999))

		err := r.catalogDB.WithContext(r.ctx).Exec(
			`INSERT INTO products (merchant_id, category_id, name, description, price, count_in_stock, brand, weight, slug_product, image_product, barcode, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			merchant.MerchantID, category.CategoryID, name,
			fmt.Sprintf("Description for %s", name),
			price, countInStock, brand, weight, slug, image, barcode, now, now,
		).Error
		if err != nil {
			r.logger.Error("Failed to create product:", zap.Error(err))
			return err
		}
	}

	r.logger.Info("Product seeding completed successfully.", zap.Int("count", 10))
	return nil
}

type orderSeeder struct {
	salesDB    *gorm.DB
	merchantDB *gorm.DB
	catalogDB  *gorm.DB
	ctx        context.Context
	logger     logger.LoggerInterface
}

func NewOrderSeeder(salesDB, merchantDB, catalogDB *gorm.DB, ctx context.Context, logger logger.LoggerInterface) *orderSeeder {
	return &orderSeeder{salesDB: salesDB, merchantDB: merchantDB, catalogDB: catalogDB, ctx: ctx, logger: logger}
}

type cashierRow struct {
	CashierID int
}

type productRow struct {
	ProductID int
	Price     int32
}

func (r *orderSeeder) Seed() error {
	var merchants []merchantRow
	err := r.merchantDB.WithContext(r.ctx).Raw(`SELECT merchant_id FROM merchants WHERE deleted_at IS NULL LIMIT 20`).Scan(&merchants).Error
	if err != nil {
		r.logger.Error("Failed to get merchants", zap.Error(err))
		return err
	}

	var cashiers []cashierRow
	err = r.merchantDB.WithContext(r.ctx).Raw(`SELECT cashier_id FROM cashiers WHERE deleted_at IS NULL LIMIT 20`).Scan(&cashiers).Error
	if err != nil {
		r.logger.Error("Failed to get cashiers", zap.Error(err))
		return err
	}

	if len(merchants) == 0 || len(cashiers) == 0 {
		r.logger.Error("No merchants or cashiers found, skipping order seeding")
		return nil
	}

	now := time.Now()
	for i := 0; i < 10; i++ {
		merchant := merchants[rand.Intn(len(merchants))]
		cashier := cashiers[rand.Intn(len(cashiers))]
		totalPrice := int64(rand.Intn(500000) + 50000)

		var orderID int
		err := r.salesDB.WithContext(r.ctx).Raw(
			`INSERT INTO orders (merchant_id, cashier_id, total_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?) RETURNING order_id`,
			merchant.MerchantID, cashier.CashierID, totalPrice, now, now,
		).Scan(&orderID).Error
		if err != nil {
			r.logger.Error("Failed to create order", zap.Error(err))
			return err
		}

		var products []productRow
		err = r.catalogDB.WithContext(r.ctx).Raw(
			`SELECT product_id, price FROM products WHERE merchant_id = ? AND deleted_at IS NULL LIMIT 10`, merchant.MerchantID,
		).Scan(&products).Error
		if err != nil {
			r.logger.Error("Failed to get products", zap.Error(err))
			return err
		}

		if len(products) == 0 {
			r.logger.Debug("No products found for merchant", zap.Int("merchant_id", merchant.MerchantID))
			continue
		}

		for j := 0; j < rand.Intn(5)+1; j++ {
			product := products[rand.Intn(len(products))]
			quantity := int32(rand.Intn(5) + 1)
			price := product.Price * quantity

			err := r.salesDB.WithContext(r.ctx).Exec(
				`INSERT INTO order_items (order_id, product_id, quantity, price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
				orderID, product.ProductID, quantity, price, now, now,
			).Error
			if err != nil {
				r.logger.Error("Failed to create order item", zap.Error(err))
				return err
			}
		}
	}

	r.logger.Info("Order seeding completed successfully.", zap.Int("count", 10))
	return nil
}

type transactionSeeder struct {
	salesDB    *gorm.DB
	merchantDB *gorm.DB
	ctx        context.Context
	logger     logger.LoggerInterface
}

func NewTransactionSeeder(salesDB, merchantDB *gorm.DB, ctx context.Context, logger logger.LoggerInterface) *transactionSeeder {
	return &transactionSeeder{salesDB: salesDB, merchantDB: merchantDB, ctx: ctx, logger: logger}
}

type orderRow struct {
	OrderID int
}

func (r *transactionSeeder) Seed() error {
	var orders []orderRow
	err := r.salesDB.WithContext(r.ctx).Raw(`SELECT order_id FROM orders WHERE deleted_at IS NULL LIMIT 20`).Scan(&orders).Error
	if err != nil {
		r.logger.Error("Failed to get orders:", zap.Error(err))
		return err
	}

	var merchants []merchantRow
	err = r.merchantDB.WithContext(r.ctx).Raw(`SELECT merchant_id FROM merchants WHERE deleted_at IS NULL LIMIT 20`).Scan(&merchants).Error
	if err != nil {
		r.logger.Error("Failed to get merchants:", zap.Error(err))
		return err
	}

	now := time.Now()
	for i := 0; i < 10; i++ {
		selectedMerchant := merchants[rand.Intn(len(merchants))]
		selectedOrder := orders[rand.Intn(len(orders))]
		amount := int32(100 + i)
		changeAmount := int32(5 + i)

		err := r.salesDB.WithContext(r.ctx).Exec(
			`INSERT INTO transactions (order_id, merchant_id, payment_method, amount, change_amount, payment_status, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			selectedOrder.OrderID, selectedMerchant.MerchantID, "Credit Card", amount, &changeAmount, "Completed", now, now,
		).Error
		if err != nil {
			r.logger.Error("Failed to create transaction:", zap.Error(err))
			return err
		}
	}

	r.logger.Info("Successfully seeded 10 transactions.", zap.Int("count", 10))
	return nil
}
