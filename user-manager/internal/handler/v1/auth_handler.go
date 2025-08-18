package v1handler

import (
	"net/http"
	v1dto "user-management-api/internal/dto/v1"
	v1services "user-management-api/internal/services/v1"
	"user-management-api/internal/utils"
	"user-management-api/internal/validation"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service v1services.AuthService
}

func NewAuthHandler(service v1services.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (ah *AuthHandler) Login(c *gin.Context) {
	var input v1dto.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {

		utils.ResponseValidator(c, validation.HandleValidationError(err))
		return
	}
	 accesstoken,refreshtoken,expiresIn, err := ah.service.Login(c, input.Email, input.Password)
	 if err != nil {
		utils.ResponseError(c, err)
		return
	}
	response := v1dto.LoginResponse{
		AccessToken: accesstoken,
		ExpiresIn: expiresIn,
		Refreshtoken:refreshtoken,
	}
	utils.ResponSuccess(c, http.StatusOK, "Login successful", response)
}
func (ah *AuthHandler) Logout(c *gin.Context) {
	utils.ResponSuccess(c, http.StatusOK, "Logout successful")
}
