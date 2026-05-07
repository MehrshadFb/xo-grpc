PROTO_DIR := api/proto
GEN_DIR := gen/go

proto:
	mkdir -p $(GEN_DIR)
	protoc -I $(PROTO_DIR) \
	  --go_out=$(GEN_DIR) --go_opt=paths=source_relative \
	  --go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative \
	  $(PROTO_DIR)/xo/v1/common.proto \
	  $(PROTO_DIR)/xo/v1/lobby.proto \
	  $(PROTO_DIR)/xo/v1/game.proto
