package product_test

import (
	"context"
	"testing"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-product/repository"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// fakeCategoryQueryRepository stands in for the category gRPC adapter. These
// repository/service suites exercise the local product repositories, so the
// cross-service dependencies never leave the process.
type fakeCategoryQueryRepository struct{}

func (fakeCategoryQueryRepository) FindById(_ context.Context, categoryID int) (*models.Category, error) {
	return &models.Category{CategoryID: int32(categoryID), Name: "Test Category"}, nil
}

func (fakeCategoryQueryRepository) FindByIds(_ context.Context, ids []int) ([]*models.Category, error) {
	categories := make([]*models.Category, 0, len(ids))
	for _, id := range ids {
		categories = append(categories, &models.Category{CategoryID: int32(id), Name: "Test Category"})
	}
	return categories, nil
}

func (fakeCategoryQueryRepository) FindByName(_ context.Context, name string) (*models.Category, error) {
	return &models.Category{CategoryID: 1, Name: name}, nil
}

type fakeMerchantQueryRepository struct{}

func (fakeMerchantQueryRepository) FindById(_ context.Context, merchantID int) (*models.Merchant, error) {
	return &models.Merchant{MerchantID: int32(merchantID), Name: "Test Merchant"}, nil
}

// newLocalRepositories wires the local product repositories with the fake
// cross-service dependencies used by these suites.
func newLocalRepositories(db *gorm.DB) *repository.Repositories {
	return &repository.Repositories{
		ProductQuery:   repository.NewProductQueryRepository(db),
		ProductCommand: repository.NewProductCommandRepository(db),
		CategoryQuery:  fakeCategoryQueryRepository{},
		MerchantQuery:  fakeMerchantQueryRepository{},
	}
}

type ProductRepositoryTestSuite struct {
	suite.Suite
	ts        *tests.TestSuite
	repo      *repository.Repositories
	productID int
}

func (s *ProductRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.ts = ts

	s.Require().NoError(err)

	productQueries := s.ts.GormDB()
	s.repo = newLocalRepositories(productQueries)

	// Seed a merchant and category for product tests
	var userID, categoryID int
	err = s.ts.GormDB().Raw(`INSERT INTO users (firstname, lastname, email, password, verification_code, is_verified) VALUES ($1, $2, $3, $4, 'test-verify', true) RETURNING user_id`,
		"Prod", "Repo", "prod.repo@example.com", "password123",
	).Scan(&userID).Error
	s.Require().NoError(err)

	err = s.ts.GormDB().Raw(`INSERT INTO categories (name, description) VALUES ($1, $2) RETURNING category_id`,
		"Test Category", "Category for product tests",
	).Scan(&categoryID).Error
	s.Require().NoError(err)

	err = s.ts.GormDB().Raw(`INSERT INTO merchants (user_id, name, description, address, contact_email, contact_phone, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING merchant_id`,
		userID, "Test Merchant", "Desc", "Addr", "pm@example.com", "123", "active",
	).Scan(&userID).Error
	s.Require().NoError(err)
}

func (s *ProductRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *ProductRepositoryTestSuite) Test1_CreateProduct() {
	ctx := context.Background()

	req := &requests.CreateProductRequest{
		MerchantID:   1,
		CategoryID:   1,
		Name:         "Test Product",
		Description:  "Product description",
		Price:        100,
		CountInStock: 50,
		Brand:        "Test Brand",
		Weight:       1,
		ImageProduct: "test.jpg",
	}

	res, err := s.repo.ProductCommand.CreateProduct(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(req.Name, res.Name)
	s.productID = int(res.ProductID)
}

func (s *ProductRepositoryTestSuite) Test2_FindById() {
	s.Require().NotZero(s.productID)
	ctx := context.Background()

	found, err := s.repo.ProductQuery.FindById(ctx, s.productID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(s.productID, int(found.ProductID))
}

func (s *ProductRepositoryTestSuite) Test3_UpdateProduct() {
	s.Require().NotZero(s.productID)
	ctx := context.Background()

	req := &requests.UpdateProductRequest{
		ProductID:    &s.productID,
		MerchantID:   1,
		CategoryID:   1,
		Name:         "Updated Product",
		Description:  "Updated description",
		Price:        200,
		CountInStock: 100,
		Brand:        "Updated Brand",
		Weight:       2,
		ImageProduct: "updated.jpg",
	}

	res, err := s.repo.ProductCommand.UpdateProduct(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal("Updated Product", res.Name)
}

func (s *ProductRepositoryTestSuite) Test4_TrashAndRestore() {
	s.Require().NotZero(s.productID)
	ctx := context.Background()

	// Trash
	trashed, err := s.repo.ProductCommand.TrashedProduct(ctx, s.productID)
	s.NoError(err)
	s.NotNil(trashed)

	// Restore
	restored, err := s.repo.ProductCommand.RestoreProduct(ctx, s.productID)
	s.NoError(err)
	s.NotNil(restored)

	// Verify restored
	found, err := s.repo.ProductQuery.FindById(ctx, s.productID)
	s.NoError(err)
	s.NotNil(found)
}

func (s *ProductRepositoryTestSuite) Test5_DeletePermanent() {
	s.Require().NotZero(s.productID)
	ctx := context.Background()

	// Must be trashed first for permanent delete
	_, err := s.repo.ProductCommand.TrashedProduct(ctx, s.productID)
	s.NoError(err)

	success, err := s.repo.ProductCommand.DeleteProductPermanent(ctx, s.productID)
	s.NoError(err)
	s.True(success)
}

func TestProductRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(ProductRepositoryTestSuite))
}
