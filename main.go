package main

import (
	"logexplorationapi/pkg/routes"
    "github.com/gin-gonic/gin"
)

func main() {

	r := routes.SetUpRouter() //initialise
	r.Run() //run server on port 8080
}
