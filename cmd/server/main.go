package main

import (
	"github.com/dev3mike/go-api-swagger-boilerplate/cmd/server/setup"
	_ "github.com/dev3mike/go-api-swagger-boilerplate/docs"
)

// @title GO Backend API Boilerplate
// @version 1.0
// @description Add your api description here
func main() {
	setup.SetupServerPrerequisites()
	setup.StartServer()
}
