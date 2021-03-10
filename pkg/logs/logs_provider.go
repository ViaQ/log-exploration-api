package logs

import "time"

type LogsProvider interface {
	FilterByIndex(index string) ([]string,error,int)
	FilterByTime(startTime time.Time,finishTime time.Time) ([]string,error,int)
	GetAllLogs() ([]string,error,int)
	FilterByPodName(podName string) ([]string,error,int)
}
