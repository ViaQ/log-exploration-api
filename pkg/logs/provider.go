package logs

type LogsProvider interface {
	FilterLogs(params Parameters) ([]string, error)
}
