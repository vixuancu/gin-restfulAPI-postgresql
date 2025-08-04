package repository

import (
	"database/sql"
	"fmt"
	"hoc-gin/internal/models"
)

type SqlUserRepository struct {
	db *sql.DB
}

func NewSqlUserRepository(DB *sql.DB) UserRepository {
	return &SqlUserRepository{db: DB}
}

func (ur *SqlUserRepository) Create(user *models.User) error {
	err := ur.db.QueryRow("insert into users (name,email) values ($1, $2) returning user_id", user.Name, user.Email).Scan(&user.Id)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (ur *SqlUserRepository) FindById(id int, user *models.User) error {
	row := ur.db.QueryRow("select * from users where user_id = $1", id)
	err := row.Scan(&user.Id, &user.Name, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user with id %d not found", id)
		}
		return fmt.Errorf("failed to find user by id: %w", err)
	}
	return nil
}
