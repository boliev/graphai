migrate:
	migrate -path migrations -database postgres://graphai:123456@localhost:5432/graphai?sslmode=disable up