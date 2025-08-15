package v1handler

import (
	"fmt"
	"net/http"
	v1dto "user-management-api/internal/dto/v1"
	v1services "user-management-api/internal/services/v1"
	"user-management-api/internal/utils"
	"user-management-api/internal/validation"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService v1services.UserService
}
type GetUserByUUIDParam struct {
	UUID string `uri:"uuid" binding:"required,uuid"`
}
type GetUsersParams struct { // sử dụng query string để tìm kiếm
	Search string `form:"search" binding:"omitempty,search"`
	Page   int    `form:"page" binding:"omitempty,min=1"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

func NewUserHandler(service v1services.UserService) *UserHandler {
	return &UserHandler{
		userService: service,
	}
}
func (uh *UserHandler) GetAllUsers(c *gin.Context) {
	var params GetUsersParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ResponseValidator(c, validation.HandleValidationError(err))
		return
	}
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 10 // mặc định là 10
	}

	utils.ResponSuccess(c, http.StatusOK, "")
}
func (uh *UserHandler) CreateUsers(c *gin.Context) {
	var input v1dto.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {

		utils.ResponseValidator(c, validation.HandleValidationError(err))
		return
	}

	user := input.MapCreateInputToModel()

	createdUser, err := uh.userService.CreateUser(c, user)
	if err != nil {
		utils.ResponseError(c, err)
		return // dừng func ở đây không nó chạy xuống success
	}
	dtoUser := v1dto.MapUserToDTO(createdUser)
	// log.Println("Start processing")
	// time.Sleep(10 * time.Second) // Giả lập thời gian xử lý lâu
	// log.Println("End processing")

	utils.ResponSuccess(c, http.StatusCreated,"User Create successfully", dtoUser)

}
func (uh *UserHandler) GetUserByUUID(c *gin.Context) {
	var param GetUserByUUIDParam
	if err := c.ShouldBindUri(&param); err != nil {
		utils.ResponseValidator(c, validation.HandleValidationError(err))
		return
	}

	utils.ResponSuccess(c, http.StatusOK, "")
}
func (uh *UserHandler) UpdateUser(c *gin.Context) {
	var param GetUserByUUIDParam
	if err := c.ShouldBindUri(&param); err != nil {
		utils.ResponseValidator(c, validation.HandleValidationError(err))
		return
	}
	userUuid, err := uuid.Parse(param.UUID)
	if err != nil {
		utils.ResponseError(c, fmt.Errorf("invalid UUID format: %v", err))
		return
	}

	var inputUpdate v1dto.UpdateUserInput
	if err := c.ShouldBindJSON(&inputUpdate); err != nil {

		utils.ResponseValidator(c, validation.HandleValidationError(err))
		return
	}
	user := inputUpdate.MapUpdateToModel(userUuid)

	updatedUser, err := uh.userService.UpdateUser(c, user)
	if err != nil {
		utils.ResponseError(c, err) // Trả về lỗi nếu có
		return                      // dừng func ở đây không nó chạy xuống ResponSuccess
	}
	userDto := v1dto.MapUserToDTO(updatedUser) // Chuyển đổi sang DTO để trả về
	utils.ResponSuccess(c, http.StatusOK,"Updated user successfully" ,userDto)
}

// xoa mem
func (uh *UserHandler) SoftDeleteUser(c *gin.Context) {
	var param GetUserByUUIDParam
	if err := c.ShouldBindUri(&param); err != nil {
		utils.ResponseValidator(c, validation.HandleValidationError(err))
		return
	}
	userUuid, err := uuid.Parse(param.UUID)
	if err != nil {
		utils.ResponseError(c, fmt.Errorf("invalid UUID format: %v", err))
		return
	}
	softDeleteUser, err := uh.userService.SoftDeleteUser(c, userUuid)
	if err != nil {
		utils.ResponseError(c, err) // Trả về lỗi nếu có
		return                      // dừng func ở đây không nó chạy xuống ResponSuccess
	}
	userDto := v1dto.MapUserToDTO(softDeleteUser)  // Chuyển đổi sang DTO để trả về
	utils.ResponSuccess(c, http.StatusOK, "User Deleted Successfully",userDto) // Trả về 204 No Content khi xóa thành công
}
func (uh *UserHandler) RestoreUser(c *gin.Context) {
	var param GetUserByUUIDParam
	if err := c.ShouldBindUri(&param); err != nil {
		utils.ResponseValidator(c, validation.HandleValidationError(err))
		return
	}
	userUuid, err := uuid.Parse(param.UUID)
	if err != nil {
		utils.ResponseError(c, fmt.Errorf("invalid UUID format: %v", err))
		return
	}
	restoreUser, err := uh.userService.RestoreUser(c, userUuid)
	if err != nil {
		utils.ResponseError(c, err) // Trả về lỗi nếu có
		return                      // dừng func ở đây không nó chạy xuống ResponSuccess
	}
	userDto := v1dto.MapUserToDTO(restoreUser)     // Chuyển đổi sang DTO để trả về
	utils.ResponSuccess(c, http.StatusOK, "restore user Successfully", userDto) // Trả về 204 No Content khi xóa thành công
}
func (uh *UserHandler) DeleteUser(c *gin.Context) {
	var param GetUserByUUIDParam
	if err := c.ShouldBindUri(&param); err != nil {
		utils.ResponseValidator(c, validation.HandleValidationError(err))
		return
	}
	userUuid, err := uuid.Parse(param.UUID)
	if err != nil {
		utils.ResponseError(c, fmt.Errorf("invalid UUID format: %v", err))
		return
	}
	err = uh.userService.DeleteUser(c, userUuid) // Xóa người dùng
	if err != nil {
		utils.ResponseError(c, err) // Trả về lỗi nếu có
		return                      // dừng func ở đây không nó chạy xuống ResponSuccess
	}
	utils.ResponseStatusCode(c, http.StatusNoContent) // Trả về 204 No Content khi xóa thành công
}

func (uh *UserHandler) PanicUser(c *gin.Context) {
	// var a []int
	// fmt.Println(a[1]) // Gây panic để kiểm tra middleware recovery
}
