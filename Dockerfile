FROM golang:1.25.6-alpine AS builder

WORKDIR /app

RUN apk add --no-cache protobuf make

COPY go.mod go.sum ./
RUN go mod download

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

COPY . .

RUN make proto
RUN go build -o xo-grpc ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/xo-grpc .

EXPOSE 50051

CMD ["./xo-grpc"]