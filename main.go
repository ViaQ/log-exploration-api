package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"logexplorationapi/pkg/configuration"
	"logexplorationapi/pkg/elastic"
	"logexplorationapi/pkg/logscontroller"
)

func main() {
	router := gin.Default()
	appConf := configuration.ParseArgs()
	repository,err := elastic.NewElasticRepository(appConf)
	if(err!=nil){
		fmt.Println(err)
		return
	}
	logscontroller.NewLogsController(repository, router)
	router.Run()
}
