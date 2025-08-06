package repository

import (
	"fmt"
	"hoc-gin/internal/models"

	"gorm.io/gorm"
)

type SqlUserRepository struct {
	db *gorm.DB
}

func NewSqlUserRepository(DB *gorm.DB) UserRepository {
	return &SqlUserRepository{
		db: DB,
	}
}

func (ur *SqlUserRepository) Create(user *models.User) error {
	if err := ur.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}
	return nil
}

func (ur *SqlUserRepository) FindById(user *models.User, id int) error {
	if err := ur.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("user with ID %d not found", id)
		}
		return err
	}
	return nil
}
