FROM golang:1.24.3-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/book_service
RUN go build -o book_service cmd/main.go
CMD ["./book_service"]
