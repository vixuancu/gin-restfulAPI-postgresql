package v1services

import (
	"user-management-api/internal/repository"
)

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		userRepo: repo,
	}
}
func (us *userService) GetAllUsers(search string, page int, limit int) {

}
func (us *userService) CreateUser() {

}
func (us *userService) GetUserByUUID(uuid string) {

}
func (us *userService) UpdateUser(uuid string) {

}
func (us *userService) DeleteUser(uuid string) {

}
