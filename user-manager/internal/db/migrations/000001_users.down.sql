-- xóa trigger 
DROP TRIGGER IF EXISTS set_user_updated_at ON users;
-- xóa function
DROP FUNCTION IF EXISTS update_user_updated_at();
-- xóa bảng users: xóa bảng sẽ xóa luôn index 
DROP TABLE IF EXISTS users;