package v1dto


type UserDTO struct {
	UUID   string `json:"uuid"`
	Name   string `json:"full_name"`
	Email  string `json:"email_address"`
	Age    int    `json:"age"`
	Status string `json:"status"`
	Level  string `json:"level"`
}
type CreateUserInput struct {
	
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email,email_advanced"`
	Password string `json:"password" binding:"required,password_strong"`
	Age      int    `json:"age" binding:"required,gt=0,lte=120"`
	Status   int    `json:"status" binding:"required,oneof=1 2"`
	Level    int    `json:"level" binding:"required,oneof=1 2"`
}
type UpdateUserInput struct {
	
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email,email_advanced"`
	Password string `json:"password" binding:"omitempty,password_strong"`
	Age      int    `json:"age" binding:"required,gt=0,lte=120"`
	Status   int    `json:"status" binding:"required,oneof=1 2"`
	Level    int    `json:"level" binding:"required,oneof=1 2"`
}
func (input *CreateUserInput) MapCreateToModel()  {
	
}
func (input *UpdateUserInput) MapUpdateToModel() {
	
}


func mapStatusText(status int) string {
	switch status {
	case 1:
		return "Show"
	case 2:
		return "Hide"
	default:
		return "None"
	}
}
func mapLevelText(level int) string {
	switch level {
	case 1:
		return "Admin"
	case 2:
		return "Member"
	default:
		return "None"
	}
}