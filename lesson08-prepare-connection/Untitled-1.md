migrate create -ext sql -dir internal/db/migrations -seq users

migrate -path internal/db/migrations -database "postgresql://vixuancu:123456@localhost:5433/master-golang?sslmode=disable" up
migrate -path internal/db/migrations -database "postgresql://vixuancu:123456@localhost:5433/master-golang?sslmode=disable" down 1