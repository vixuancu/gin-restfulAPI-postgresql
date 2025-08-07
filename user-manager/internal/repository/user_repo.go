package repository



type SqlUserRepository struct {
}

// constructor function to create a new UserRepository instance //làm một hàm khởi tạo để tạo một instance mới của UserRepository
func NewSqlUserRepository() UserRepository {
	return &SqlUserRepository{}
}
func (repo *SqlUserRepository) FindAll()  {
	
}
func (repo *SqlUserRepository) Create()  {

}
func (repo *SqlUserRepository) FindByUUID(uuid string) {
	
	
}
func (repo *SqlUserRepository) Update(uuid string)  {
	
}
func (repo *SqlUserRepository) Delete(uuid string)  {
	

}
func (repo *SqlUserRepository) FindByEmail(email string) {
	
}