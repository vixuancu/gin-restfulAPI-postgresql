importdb:
	docker exec -i postgres-db psql -U vixuancu -d master-golang < ./backupdb-master-golang.sql
exportdb:
	docker exec -i postgres-db pg_dump -U vixuancu -d master-golang > ./backupdb-master-golang.sql
server:
	go run .
start-container:
	docker-compose up -d
remove-container:
	docker-compose down