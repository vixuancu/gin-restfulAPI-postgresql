ALTER TABLE users RENAME COLUMN email TO user_email;
ALTER TABLE users ALTER COLUMN user_email SET data type text;