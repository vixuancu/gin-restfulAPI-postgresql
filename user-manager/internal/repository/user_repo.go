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
func (ur *SqlUserRepository) GetAll(ctx context.Context, search, orderBy, sort string, limit, offset int32) ([]sqlc.User, error) {
	var users []sqlc.User
	var err error
	switch {
	case orderBy == "user_id" && sort == "asc":
		users,err =ur.db.ListUsersIdAsc(ctx, sqlc.ListUsersIdAscParams{
			Limit:  limit,
			Offset: offset,
			Search: search,
		})
	case orderBy == "user_id" && sort == "desc":
		users,err =ur.db.ListUsersIdDesc(ctx, sqlc.ListUsersIdDescParams{
			Limit:  limit,
			Offset: offset,
			Search: search,
		})
	case orderBy == "user_created_at" && sort == "asc":
		users,err =ur.db.ListUsersCreateAtAsc(ctx, sqlc.ListUsersCreateAtAscParams{
			Limit:  limit,
			Offset: offset,
			Search: search,
		})
	case orderBy == "user_created_at" && sort == "desc":
		users,err =ur.db.ListUsersCreateAtDesc(ctx, sqlc.ListUsersCreateAtDescParams{
			Limit:  limit,
			Offset: offset,
			Search: search,
		})
	}
	if err != nil {
		return []sqlc.User{}, err
	}
	return users, nil
}

func (ur *SqlUserRepository) CountUsers(ctx context.Context, search string) (int64, error) {
	total,err := ur.db.CountUsers(ctx,search)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (ur *SqlUserRepository) Create(ctx context.Context, userParams sqlc.CreateUserParams) (sqlc.User, error) {
	log.Printf("%+v", userParams) // %+v → in giá trị mặc định + tên trường (struct)
	log.Printf("%+v", ur.db)      // in giá trị mặc định + tên trường (struct)
	user, err := ur.db.CreateUser(ctx, userParams)
	if err != nil {
		return sqlc.User{}, err
	}
	return user, nil
}
func (ur *SqlUserRepository) FindByUUID(uuid string) {

}
func (ur *SqlUserRepository) Update(ctx context.Context, input sqlc.UpdateUserParams) (sqlc.User, error) {
	user, err := ur.db.UpdateUser(ctx, input) // Cập nhật thông tin người dùng
	if err != nil {
		return sqlc.User{}, err
	}
	return user, nil
}

func (ur *SqlUserRepository) Delete(ctx context.Context, uuid uuid.UUID) (sqlc.User, error) {
	user, err := ur.db.TrashUser(ctx, uuid) // Thực hiện xóa mềm người dùng
	if err != nil {
		return sqlc.User{}, err
	}
	return user, nil

}
func (ur *SqlUserRepository) SoftDelete(ctx context.Context, uuid uuid.UUID) (sqlc.User, error) {
	user, err := ur.db.SoftDeleteUser(ctx, uuid) // Thực hiện xóa mềm người dùng
	if err != nil {
		return sqlc.User{}, err
	}
	return user, nil

}
func (ur *SqlUserRepository) Restore(ctx context.Context, uuid uuid.UUID) (sqlc.User, error) {

	user, err := ur.db.RestoreUser(ctx, uuid) // Thực hiện xóa mềm người dùng
	if err != nil {
		return sqlc.User{}, err
	}
	return user, nil

}
func (ur *SqlUserRepository) FindByEmail(email string) {

}
