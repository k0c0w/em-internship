BINARY_NAME=effective-mobile

.PHONY: generate
generate:
	oapi-codegen -config swagger/oapi-codegen.config.yaml swagger/swagger.yml
	copy swagger\swagger.yml internal\http\api\swagger\swagger.yml

.PHONY: build
build: generate
	go build -o $(BINARY_NAME) cmd/main.go