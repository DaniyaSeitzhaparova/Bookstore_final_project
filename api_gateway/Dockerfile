FROM golang:1.24.3-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/api_gateway
RUN go build -o api_gateway cmd/main.go
CMD ["./api_gateway"]
