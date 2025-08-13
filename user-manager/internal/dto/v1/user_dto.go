package v1dto

import (
	"user-management-api/internal/db/sqlc"
	"user-management-api/internal/utils"
)

type UserDTO struct {
	UUID      string `json:"uuid"`
	Name      string `json:"full_name"`
	Email     string `json:"email_address"`
	Age       *int    `json:"age"`
	Status    string `json:"status"`
	Level     string `json:"level"`
	CreatedAt string `json:"created_at"`
}
type CreateUserInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email,email_advanced"`
	Password string `json:"password" binding:"required,password_strong"`
	Age      int    `json:"age" binding:"omitempty,gt=0,lte=120"`
	Status   int    `json:"status" binding:"required,oneof=1 2 3"`
	Level    int    `json:"level" binding:"required,oneof=1 2 3"`
}
type UpdateUserInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email,email_advanced"`
	Password string `json:"password" binding:"omitempty,password_strong"`
	Age      int    `json:"age" binding:"omitempty,gt=0,lte=120"`
	Status   int    `json:"status" binding:"required,oneof=1 2"`
	Level    int    `json:"level" binding:"required,oneof=1 2"`
}

func (input *CreateUserInput) MapCreateInputToModel() sqlc.CreateUserParams {
	return sqlc.CreateUserParams{
		UserEmail:    input.Email,
		UserPassword: input.Password,
		UserFullname: input.Name,
		UserAge:      utils.ConvertToInt32Pointer(input.Age),
		UserStatus:   int32(input.Status),
		UserLevel:    int32(input.Level),
	}

}
func (input *UpdateUserInput) MapUpdateToModel() {

}

func MapUserToDTO(user sqlc.User) *UserDTO {
	dto := &UserDTO{
		UUID:      user.Uuid.String(), // .String() vì uuid có method String() để chuyển đổi thành string
		Name:      user.UserFullname,
		Email:     user.UserEmail,
		Status:    mapStatusText(int(user.UserStatus)),
		Level:     mapLevelText(int(user.UserLevel)),
		CreatedAt: user.UserCreatedAt.Format("2006-01-02 15:04:05"),
	}
	if user.UserAge != nil {
		age := int(*user.UserAge)
		dto.Age = &age // Chuyển đổi int thành *int
	}
	// if user.UserDeletedAt.Valid{
	// 	dto.DeletedAt = user.UserDeletedAt.Time.Format("2006-01-02 15:04:05")
	// }else {
	// 	dto.DeletedAt = "" // Nếu không có giá trị thì để trống
	// }
	return dto
}

func mapStatusText(status int) string {
	switch status {
	case 1:
		return "Active"
	case 2:
		return "Inactive"
	case 3:
		return "Banned"
	default:
		return "Unknown"
	}
}
func mapLevelText(level int) string {
	switch level {
	case 1:
		return "Admin"
	case 2:
		return "moder"
	case 3:
		return "Member"
	default:
		return "Unknown"
	}
}
