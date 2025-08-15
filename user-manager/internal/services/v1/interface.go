package v1services

import (
	"user-management-api/internal/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserService interface {
	GetAllUsers(search string, page int, limit int)
	CreateUser(ctx *gin.Context, input sqlc.CreateUserParams) (sqlc.User, error)
	GetUserByUUID(uuid string)
	UpdateUser(ctx *gin.Context, input sqlc.UpdateUserParams) (sqlc.User, error)
	SoftDeleteUser(c *gin.Context, uuid uuid.UUID) (sqlc.User, error)
	RestoreUser(c *gin.Context, uuid uuid.UUID) (sqlc.User, error)
	DeleteUser(c *gin.Context, uuid uuid.UUID)  error
}
