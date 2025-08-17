package v1services

import (
	"user-management-api/internal/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserService interface {
	GetAllUsers(ctx *gin.Context, search, orderBy, sort string, limit,page  int32,deleted bool) ([]sqlc.User,int32, error)
	CreateUser(ctx *gin.Context, input sqlc.CreateUserParams) (sqlc.User, error)
	GetUserByUUID(c *gin.Context, uuid uuid.UUID) (sqlc.User, error)
	UpdateUser(ctx *gin.Context, input sqlc.UpdateUserParams) (sqlc.User, error)
	SoftDeleteUser(c *gin.Context, uuid uuid.UUID) (sqlc.User, error)
	RestoreUser(c *gin.Context, uuid uuid.UUID) (sqlc.User, error)
	DeleteUser(c *gin.Context, uuid uuid.UUID)  error
}
type AuthService interface {
	Login(c *gin.Context,email,password string) error
	Logout(c *gin.Context) error
}
