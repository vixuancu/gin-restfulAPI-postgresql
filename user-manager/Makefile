include .env
export
MIGRATION_DIRS = internal/db/migrations
CONN_STRING = postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
ENV_FILE = .env
PROD_COMPOSE= docker-compose.prod.yml
NOAPP_COMPOSE= docker-compose.noapp.yml
DEV_COMPOSE= docker-compose.dev.yml
# import database
importdb:
	docker exec -i postgres-db psql -U vixuancu -d master-golang < ./backupdb-master-golang.sql
# export database
exportdb:
	docker exec -i postgres-db pg_dump -U vixuancu -d master-golang > ./backupdb-master-golang.sql
# Build binary file
build:
	go build -o bin/myapp ./cmd/api
# Run binary file
run-binary:
	./bin/myapp
# run server
server:
	go run ./cmd/api
# generate sqlc
sqlc:
	sqlc generate
# start and remove docker container
start-container:
	docker-compose up -d
remove-container:
	docker-compose down
# migrate database
# create a new migration (make migrate-create NAME=profiles)
migrate-create: 
	migrate create -ext sql -dir $(MIGRATION_DIRS) -seq $(NAME)
# run all pending migrations (make migrate-up)
migrate-up:
	migrate -path $(MIGRATION_DIRS) -database "$(CONN_STRING)" up
# rollback the last migration (make migrate-down)
migrate-down:
	migrate -path $(MIGRATION_DIRS) -database "$(CONN_STRING)" down 1
# rollback n versions  (make migrate-down-n N=2)
migrate-down-n:
	migrate -path $(MIGRATION_DIRS) -database "$(CONN_STRING)" down $(N)
# trở về phiên bản cụ thể (make migrate-force VERSION=1)
migrate-force:
	migrate -path $(MIGRATION_DIRS) -database "$(CONN_STRING)" force $(VERSION)
# drop everything (include schema and data) (make migrate-drop)
migrate-drop:
	migrate -path $(MIGRATION_DIRS) -database "$(CONN_STRING)" drop
# apple migrations to a specific version (make migrate-goto VERSION=1) - goto này là chạy đến bao gồm phiên bản trước luôn
migrate-goto:
	migrate -path $(MIGRATION_DIRS) -database "$(CONN_STRING)" goto $(VERSION)
# Build container for noapp
noapp:
	docker-compose -f $(NOAPP_COMPOSE) down
	docker-compose -f $(NOAPP_COMPOSE) --env-file $(ENV_FILE) up -d --build
# Stop All container noapp
stop-noapp:
	docker-compose -f $(NOAPP_COMPOSE) down
# Build container for development
dev:
	docker-compose -f $(DEV_COMPOSE) down
	docker-compose -f $(DEV_COMPOSE) --env-file $(ENV_FILE) up  --build
# Stop All container development
stop-dev:
	docker-compose -f $(DEV_COMPOSE) down
# Build container for production
prod:
	docker-compose -f $(PROD_COMPOSE) down
	docker-compose -f $(PROD_COMPOSE) --env-file $(ENV_FILE) up -d --build
# Stop All container production
stop-prod:
	docker-compose -f $(PROD_COMPOSE) down
# View Logs container production
logs-prod:
	docker-compose -f $(PROD_COMPOSE) logs -f --tail=100
# Open bash in container golang-api
bash:
	docker exec -it golang-api /bin/sh
# cấu trúc dự án
tree:
	tree
.PHONY: importdb exportdb server start-container remove-container migrate-create migrate-up migrate-down migrate-force migrate-drop migrate-goto migrate-down-n sqlc build run-binary prod stop-prod logs-prod bash noapp stop-noapp dev stop-dev tree