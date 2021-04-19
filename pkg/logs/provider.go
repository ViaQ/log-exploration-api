package logs

type LogsProvider interface {
	FilterLogs(index string, podname string, namespace string,
		starttime string, finishtime string, level string, maxlogs string) ([]string, error)
}
