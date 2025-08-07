package v1services



type UserService interface {
	GetAllUsers(search string, page int, limit int) 
	CreateUser() 
	GetUserByUUID(uuid string ) 
	UpdateUser(uuid string) 
	DeleteUser(uuid string) 
}