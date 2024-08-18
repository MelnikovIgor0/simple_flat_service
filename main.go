package main

import (
	"bootcamp_task/server"
	_ "github.com/gofiber/swagger"
)

func main() {
	server.BuildServerAndEnv().Run()
}
