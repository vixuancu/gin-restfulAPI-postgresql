-- name: GetUser :one
select * from users where uuid = $1;


-- name: CreateUser :one
insert into users (name, email) values ($1, $2) returning *;