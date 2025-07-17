FROM golang:1.24.1

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

RUN oapi-codegen -config swagger/oapi-codegen.config.yaml swagger/swagger.yml

RUN go build -o main ./cmd/main.go

EXPOSE 8080

CMD ["./main"]