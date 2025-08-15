package repository

import (
	"context"
	"log"
	"user-management-api/internal/db/sqlc"

	"github.com/google/uuid"
)



type SqlUserRepository struct {
	db sqlc.Querier
}

// constructor function to create a new UserRepository instance //làm một hàm khởi tạo để tạo một instance mới của UserRepository
func NewSqlUserRepository(db sqlc.Querier) UserRepository {
	return &SqlUserRepository{
		db: db,
	}
}
func (ur *SqlUserRepository) FindAll()  {
	
}
func (ur *SqlUserRepository) Create(ctx context.Context, userParams sqlc.CreateUserParams) (sqlc.User, error)  {
	log.Printf("%+v", userParams) // %+v → in giá trị mặc định + tên trường (struct)
	log.Printf("%+v", ur.db) // in giá trị mặc định + tên trường (struct)
	user,err:=ur.db.CreateUser(ctx, userParams)
	if err != nil {
		return sqlc.User{}, err
	}
	return user, nil
}
func (ur *SqlUserRepository) FindByUUID(uuid string) {
	
	
}
func (ur *SqlUserRepository) Update(ctx context.Context, input sqlc.UpdateUserParams) (sqlc.User, error)  {
		user,err :=ur.db.UpdateUser(ctx, input) // Cập nhật thông tin người dùng
		if err != nil {
			return sqlc.User{}, err
		}
		return user, nil
}

	
func (ur *SqlUserRepository) Delete(ctx context.Context, uuid uuid.UUID) (sqlc.User, error) {
	user,err :=ur.db.TrashUser(ctx, uuid) // Thực hiện xóa mềm người dùng
	if err != nil {
		return sqlc.User{}, err
	}
	return user, nil

}
func (ur *SqlUserRepository) SoftDelete(ctx context.Context, uuid uuid.UUID) (sqlc.User, error) {
	user,err :=ur.db.SoftDeleteUser(ctx, uuid) // Thực hiện xóa mềm người dùng
	if err != nil {
		return sqlc.User{}, err
	}
	return user, nil

}
func (ur *SqlUserRepository) Restore(ctx context.Context, uuid uuid.UUID) (sqlc.User, error) {
	
user,err :=ur.db.RestoreUser(ctx, uuid) // Thực hiện xóa mềm người dùng
	if err != nil {
		return sqlc.User{}, err
	}
	return user, nil

}
func (ur *SqlUserRepository) FindByEmail(email string) {
	
}