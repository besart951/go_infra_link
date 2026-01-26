//go:generate go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g cmd/app/main.go -d ../../ -o ../../docs --parseDependency --parseInternal
package main

// @title Go Infra Link API
// @version 1.0
// @description API documentation for Go Infra Link.
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
