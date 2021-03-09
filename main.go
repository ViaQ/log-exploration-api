package main

import (
	"logexplorationapi/pkg1/elastic"
	"logexplorationapi/pkg1/logscontroller"
)

func main() {
	repository := elastic.NewElasticRepository()
	router := logscontroller.NewLogsController(repository)
	router.Run()
}
