ALTER TABLE profiles DROP constraint if exists fk_user;
ALTER TABLE profiles DROP COLUMN user_id;