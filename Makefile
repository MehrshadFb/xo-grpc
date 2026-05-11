PROTO_DIR := api/proto
GEN_DIR := gen/go

.PHONY: proto test race run docker-build docker-up docker-down

proto:
	mkdir -p $(GEN_DIR)
	protoc -I $(PROTO_DIR) \
	  --go_out=$(GEN_DIR) --go_opt=paths=source_relative \
	  --go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative \
	  $(PROTO_DIR)/xo/v1/common.proto \
	  $(PROTO_DIR)/xo/v1/lobby.proto \
	  $(PROTO_DIR)/xo/v1/game.proto \
	  $(PROTO_DIR)/xo/v1/health.proto

test:
	go test ./...

race:
	go test -race ./...

run:
	go run ./cmd/server

docker-build:
	docker build -t xo-grpc .

docker-up:
	docker compose up --build

docker-down:
	docker compose down