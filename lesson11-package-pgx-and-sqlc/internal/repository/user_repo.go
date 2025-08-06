package repository

import (
	"context"
	"fmt"
	"hoc-gin/internal/db/sqlc"

	"github.com/google/uuid"
)

type SqlUserRepository struct {
	db sqlc.Querier // Sử dụng interface Querier để tương tác với cơ sở dữ liệu nên không cần con trỏ
}

func NewSqlUserRepository(DB sqlc.Querier) UserRepository {
	return &SqlUserRepository{
		db: DB,
	}
}

func (ur *SqlUserRepository) Create(ctx context.Context, input sqlc.CreateUserParams) (sqlc.User, error) {
	user, err := ur.db.CreateUser(ctx, input)
	if err != nil {
		return sqlc.User{}, fmt.Errorf("failed to create user: %v", err)
	}
	return user, nil
}

func (ur *SqlUserRepository) FindById(ctx context.Context, uuid uuid.UUID) (sqlc.User, error) {
	user, err := ur.db.GetUser(ctx, uuid)
	if err != nil {
		return sqlc.User{}, fmt.Errorf("user not found by ID: %v", err)
	}
	return user, nil
}
