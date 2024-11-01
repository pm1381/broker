FROM golang:1.19-alpine

WORKDIR /broker/

# COPY go.mod and download the dependencies
COPY go.* ./
RUN go mod download
# COPY All things inside the project and build

# RUN apt-get update
# RUN apt install -y protobuf-compiler

# RUN GO111MODULE=on \
#         go get google.golang.org/protobuf/cmd/protoc-gen-go@v1.28 \
#         google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

# RUN go build -o app ./server
# CMD ["/app/app"]
COPY . .
EXPOSE 80
