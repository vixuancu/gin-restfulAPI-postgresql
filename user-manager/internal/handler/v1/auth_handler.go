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
	accesstoken, refreshtoken, expiresIn, err := ah.service.Login(c, input.Email, input.Password)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	response := v1dto.LoginResponse{
		AccessToken:  accesstoken,
		ExpiresIn:    expiresIn,
		Refreshtoken: refreshtoken,
	}
	utils.ResponSuccess(c, http.StatusOK, "Login successful", response)
}

func (ah *AuthHandler) RefreshToken(c *gin.Context) {
	var input v1dto.RefreshTokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidator(c, validation.HandleValidationError(err))
		return
	}
	accesstoken, refreshtoken, expiresIn, err := ah.service.RefreshToken(c, input.RefreshToen)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	response := v1dto.LoginResponse{
		AccessToken:  accesstoken,
		ExpiresIn:    expiresIn,
		Refreshtoken: refreshtoken,
	}
	utils.ResponSuccess(c, http.StatusOK, "refresh token generate successful", response)
}
func (ah *AuthHandler) Logout(c *gin.Context) {
	var input v1dto.RefreshTokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidator(c, validation.HandleValidationError(err))
		return
	}
	if err := ah.service.Logout(c, input.RefreshToen); err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponSuccess(c, http.StatusOK, "Logout successful")
}

func (ah *AuthHandler) ForgotPassword(c *gin.Context) {
	var input v1dto.ForgotPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidator(c, validation.HandleValidationError(err))
		return
	}
	if err:= ah.service.ForgotPassword(c, input.Email); err != nil {
		utils.ResponseError(c, err)
		return
	}
	utils.ResponSuccess(c, http.StatusOK, "ResetLink sent to email successfully")
}

func (ah *AuthHandler) ResetPassword(c *gin.Context) {
	var input v1dto.ResetPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidator(c, validation.HandleValidationError(err))
		return
	}

	if err:= ah.service.ResetPassword(c, input.Token,input.NewPassword); err != nil {
		utils.ResponseError(c, err)
		return
	}
	utils.ResponSuccess(c, http.StatusOK, "Password Reset successfully")
}
