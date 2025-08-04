-- create a new database
create database hocgolang;

-- drop a database
drop database test;

--create a new schema
create schema school;

-- drop a schema
drop schema school cascade;

-- one-one
-- Create user table
create table if not exists users (
	user_id serial primary key ,
	name varchar(50) not null,
	email varchar(100) unique not null
);
-- tạo bảng Profile
create table if not exists profiles (
	profile_id serial primary key,
	user_id int unique not null ,
	phone varchar(10),
	address varchar(100),
	constraint fk_user foreign key (user_id) references users(user_id) on delete cascade 
);

-- drop 
drop table if exists profiles;
drop table if exists users;
-- one to Many 
-- create Categories table 
create table if not exists categories (
	category_id serial primary key ,
	name varchar(50) not null
);
-- create products table
create table if not exists products (
	product_id serial primary key,
	category_id int not null,
	name varchar(100) not null,
	price int  not null check (price > 0),
	image varchar(255) ,
	status int not null check (status in (1,2)),
	constraint fk_category foreign key (category_id) references categories(category_id) on delete restrict
)

-- drop 
drop table if exists products;
drop table if exists categories;

--many to many 
-- Create students table 
create table if not exists students (
	student_id serial primary key,
	name varchar(50) not null
);
-- Create courses table 
create table if not exists courses (
	course_id serial primary key,
	name varchar(50) not null
);
-- Create students_courses table
create table if not exists students_courses (
	student_id int not null,
	course_id int not null,
	primary key (student_id,course_id),
	constraint fk_student foreign key (student_id) references students(student_id) on delete cascade,
	constraint fk_course foreign key (course_id) references courses(course_id) on delete cascade
	
);


-- drop 
drop table if exists students_courses;
drop table if exists students;
drop table if exists courses;


--------------Các truy ván SQL hay dùng ----------------------
-- Thêm dữ liệu insert into table (col1,col2) values (val1,val2)
insert into users (name,email) values ('vixuancu','vixuancu2004@gmail.com');
insert into users (name,email) values ('vixuancu1','vixuancu22004@gmail.com');

insert into profiles (user_id,address,phone) values (1,'hoang mai','0395298837');
insert into profiles (user_id,address,phone) values (2,'trung van','0395298837');

insert into categories (name) values ('Dien Thoai') , ('laptop');

insert into products (category_id,name,price,image ,status ) values 
(3,'iPhone18 promax',35000000,'image/iPhone18-promax.png',1),
(3,'iPhone17 promax',30000000,'image/iPhone17-promax.png',1);
--Cập nhật dữ liệu : update table set col1 = val1, col2 = val2 where condition
update users set email = 'test@gmail.com' where user_id = 2;

update profiles set phone = 'testphone' where user_id = 2 ;

-- Xóa dữ liệu : delete from table where condition
delete from users where user_id = 2;

delete from categories  where category_id in (1,2) ;

delete from products where category_id = 3;
delete from categories  where category_id = 3 ;


-- lấy dữ liệu : select * from table where condition on order by col [desc|asc] limit ... offset ...










