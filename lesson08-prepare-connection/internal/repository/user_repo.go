package repository

import (
	"hoc-gin/internal/models"
	"log"
)

type SqlUserRepository struct {
}

func NewSqlUserRepository() UserRepository {
	return &SqlUserRepository{}
}

func (ur *SqlUserRepository) Create(user *models.User) {
	log.Println("Creating user in SQL database")
}

func (ur *SqlUserRepository) FindById(id int) {
	log.Println("Finding user by ID in SQL database:", id)
}