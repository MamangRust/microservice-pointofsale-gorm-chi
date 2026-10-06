package merchant_test

import (
	"context"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"net"
	"testing"

	"github.com/MamangRust/microservice-point-of-sale-merchant/repository"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MerchantDocumentUpdateTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	repo       *repository.Repositories
	userID     int
	merchantID int
	docA       *models.MerchantDocument
	docB       *models.MerchantDocument
}

func (s *MerchantDocumentUpdateTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.ts = ts

	s.Require().NoError(err)

	// Placeholder gRPC listener to satisfy NewRepositories
	userLis, _ := net.Listen("tcp", "localhost:0")
	userConn, _ := grpc.NewClient(userLis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	merchantQueries := s.ts.GormDB()
	s.repo = repository.NewRepositories(merchantQueries,
		pbuser.NewUserQueryServiceClient(userConn),
		pbuser.NewUserCommandServiceClient(userConn),
	)

	err = s.ts.GormDB().Raw(`INSERT INTO users (firstname, lastname, email, password, verification_code, is_verified) VALUES ($1, $2, $3, $4, 'doc-verify', true) RETURNING user_id`,
		"Document", "Owner", "merchant.document@example.com", "password123",
	).Scan(&s.userID).Error
	s.Require().NoError(err)

	ctx := context.Background()

	merchant, err := s.repo.MerchantCommand.CreateMerchant(ctx, &requests.CreateMerchantRequest{
		UserID:       s.userID,
		Name:         "Document Merchant",
		Description:  "Doc",
		Address:      "Addr",
		ContactEmail: "doc@m.com",
		ContactPhone: "123",
		Status:       "active",
	})
	s.Require().NoError(err)
	s.merchantID = int(merchant.MerchantID)

	s.docA, err = s.repo.MerchantDocumentCommand.CreateMerchantDocument(ctx, &requests.CreateMerchantDocumentRequest{
		MerchantID:   s.merchantID,
		DocumentType: "doc_a_type",
		DocumentUrl:  "https://example.com/a.pdf",
	})
	s.Require().NoError(err)

	s.docB, err = s.repo.MerchantDocumentCommand.CreateMerchantDocument(ctx, &requests.CreateMerchantDocumentRequest{
		MerchantID:   s.merchantID,
		DocumentType: "doc_b_type",
		DocumentUrl:  "https://example.com/b.pdf",
	})
	s.Require().NoError(err)

	s.Require().NotEqual(s.docA.DocumentID, s.docB.DocumentID)
	s.Require().NotEqual(int(s.docB.DocumentID), s.merchantID,
		"fixture requires the target document ID to differ from the merchant ID")
}

func (s *MerchantDocumentUpdateTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *MerchantDocumentUpdateTestSuite) Test1_UpdateMerchantDocument_UsesDocumentID() {
	ctx := context.Background()
	docBID := int(s.docB.DocumentID)

	updated, err := s.repo.MerchantDocumentCommand.UpdateMerchantDocument(ctx, &requests.UpdateMerchantDocumentRequest{
		DocumentID:   &docBID,
		MerchantID:   s.merchantID,
		DocumentType: "doc_b_updated",
		DocumentUrl:  "https://example.com/b-updated.pdf",
		Status:       "verified",
		Note:         "Approved",
	})
	s.Require().NoError(err)
	s.Require().NotNil(updated)

	s.Equal(s.docB.DocumentID, updated.DocumentID)
	s.Equal("doc_b_updated", updated.DocumentType)
	s.Equal("verified", updated.Status)

	docA, err := s.repo.MerchantDocumentQuery.FindById(ctx, int(s.docA.DocumentID))
	s.Require().NoError(err)
	s.Require().NotNil(docA)
	s.Equal("doc_a_type", docA.DocumentType)
	s.NotEqual("doc_b_updated", docA.DocumentType)
}

func (s *MerchantDocumentUpdateTestSuite) Test2_UpdateMerchantDocumentStatus_UsesDocumentID() {
	ctx := context.Background()
	docBID := int(s.docB.DocumentID)

	updated, err := s.repo.MerchantDocumentCommand.UpdateMerchantDocumentStatus(ctx, &requests.UpdateMerchantDocumentStatusRequest{
		DocumentID: &docBID,
		MerchantID: s.merchantID,
		Status:     "rejected",
		Note:       "Invalid document",
	})
	s.Require().NoError(err)
	s.Require().NotNil(updated)

	s.Equal(s.docB.DocumentID, updated.DocumentID)
	s.Equal("rejected", updated.Status)

	docA, err := s.repo.MerchantDocumentQuery.FindById(ctx, int(s.docA.DocumentID))
	s.Require().NoError(err)
	s.Require().NotNil(docA)
	s.NotEqual("rejected", docA.Status)
}

func TestMerchantDocumentUpdateSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantDocumentUpdateTestSuite))
}
