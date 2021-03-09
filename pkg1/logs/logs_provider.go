package logs

type LogsProvider interface {
	FilterByIndex(index string) []string
	FilterByTime(startTime string, endTime string) []string
	GetAllLogs() []string
	FilterByPodName(podName string) []string
}
