package logs

import "time"

type LogsProvider interface {
	FilterByIndex(index string) ([]string, error)
	FilterByTime(startTime time.Time, finishTime time.Time) ([]string, error)
	GetAllLogs() ([]string, error)
	FilterByPodName(podName string) ([]string, error)
	FilterLogsMultipleParameters(podName string, namespace string, startTime time.Time, finishTime time.Time) ([]string, error)
	FilterLogs(params Parameters) ([]string, error)
}
