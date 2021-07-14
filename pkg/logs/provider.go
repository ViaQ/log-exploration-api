package logs

type LogsProvider interface {
	FilterLogs(params Parameters) ([]string, error)
	FilterContainerLogs(params Parameters) ([]string, error)
	FilterLabelLogs(params Parameters, labelList []string) ([]string, error)
	FilterNamespaceLogs(params Parameters) ([]string, error)
	FilterPodLogs(params Parameters) ([]string, error)
	Logs(params Parameters) ([]string, error)
	CheckReadiness() bool
}
