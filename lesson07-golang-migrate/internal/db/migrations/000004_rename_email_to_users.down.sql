ALTER TABLE users RENAME COLUMN user_email TO email;
ALTER TABLE users ALTER COLUMN email SET data type varchar(100);