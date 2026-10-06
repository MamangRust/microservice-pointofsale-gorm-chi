package handler

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-role/service"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	role_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/role_errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

type userRoleHandleGrpc struct {
	pbuserrole.UnimplementedUserRoleServiceServer

	roleQuery   service.RoleQueryService
	roleCommand service.RoleCommandService
}

func NewUserRoleHandleGrpc(query service.RoleQueryService, command service.RoleCommandService) *userRoleHandleGrpc {
	return &userRoleHandleGrpc{
		roleQuery:   query,
		roleCommand: command,
	}
}

func (s *userRoleHandleGrpc) FindByUserId(ctx context.Context, req *pbuserrole.FindByIdUserRoleRequest) (*pbrole.ApiResponsesRole, error) {
	userID := int(req.GetUserId())
	if userID <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}
	roles, err := s.roleQuery.FindByUserId(ctx, userID)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pbrole.ApiResponsesRole{Status: "success", Message: "Successfully fetched role by user ID", Data: mapRoleModels(roles)}, nil
}

func (s *userRoleHandleGrpc) AssignRoleToUser(ctx context.Context, req *pbuserrole.AssignRoleToUserRequest) (*pbuserrole.ApiResponseUserRole, error) {
	request := &requests.CreateUserRoleRequest{
		UserId: int(req.GetUserId()),
		RoleId: int(req.GetRoleId()),
	}

	userRole, err := s.roleCommand.AssignRoleToUser(ctx, request)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pbuserrole.ApiResponseUserRole{
		Status:  "success",
		Message: "Successfully assigned role to user",
		Data:    mapUserRoleModel(userRole),
	}, nil
}

func (s *userRoleHandleGrpc) RemoveRoleFromUser(ctx context.Context, req *pbuserrole.RemoveRoleFromUserRequest) (*emptypb.Empty, error) {
	request := &requests.RemoveUserRoleRequest{
		UserId: int(req.GetUserId()),
		RoleId: int(req.GetRoleId()),
	}

	if err := s.roleCommand.RemoveRoleFromUser(ctx, request); err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &emptypb.Empty{}, nil
}

func mapUserRoleModel(ur *models.UserRole) *pbuserrole.UserRoleResponse {
	if ur == nil {
		return nil
	}
	return &pbuserrole.UserRoleResponse{
		UserRoleId: ur.UserRoleID,
		UserId:     ur.UserID,
		RoleId:     ur.RoleID,
	}
}
