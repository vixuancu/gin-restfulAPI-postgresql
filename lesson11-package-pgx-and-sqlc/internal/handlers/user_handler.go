package handlers

import (
	"hoc-gin/internal/db/sqlc"
	"hoc-gin/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserResponse struct {
	UserID    int32  `json:"user_id"`
	Uuid      string `json:"uuid"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}
type UserHandler struct {
	repo repository.UserRepository // gọi interface thay vì struct cụ thể
}

func NewUserHandler(repo repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func (uh *UserHandler) GetUserById(ctx *gin.Context) {
	id := ctx.Param("uuid")
	uuid, err := uuid.Parse(id)
	if err != nil {
		if err.Error() == "no rows in result set" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user, err := uh.repo.FindById(ctx, uuid)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}
func (uh *UserHandler) CreateUser(ctx *gin.Context) {
	var input sqlc.CreateUserParams
	if err := ctx.ShouldBindJSON(&input); err != nil { // bắt buộc phải truyền & - địa chỉ của user
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user data"})
		return
	}
	user, err := uh.repo.Create(ctx, input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	response := UserResponse{
		UserID:    user.UserID,
		Uuid:      user.Uuid.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02"),
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"data": response,
	})
}
