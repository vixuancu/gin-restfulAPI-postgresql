package handlers

import (
	"errors"
	"hoc-gin/internal/models"
	"hoc-gin/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserHandler struct {
	repo repository.UserRepository // gọi interface thay vì struct cụ thể
}

func NewUserHandler(repo repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func (uh *UserHandler) GetUserById(ctx *gin.Context) {
	var user models.User
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	if err := uh.repo.FindById(&user, id); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": user, // trả về dữ liệu user
	})
}
func (uh *UserHandler) CreateUser(ctx *gin.Context) {
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil { // bắt buộc phải truyền & - địa chỉ của user
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user data"})
		return
	}
	if err := uh.repo.Create(&user); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Mã lỗi 23505 là lỗi trùng lặp khóa duy nhất
			ctx.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"data": user,
	})
}
