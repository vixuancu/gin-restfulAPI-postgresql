package repository

import "hoc-gin/internal/models"

type UserRepository interface {
	Create(user *models.User)
	FindById(id int)
}
