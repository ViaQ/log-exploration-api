package logs

type Parameters struct {
	Namespace  string `form:"namespace"`
	Index      string `form:"index"`
	Podname    string `form:"podname"`
	StartTime  string `form:"starttime"`
	FinishTime string `form:"finishtime"`
	Level      string `form:"level"`
	MaxLogs    string `form:"maxlogs"`
	ContainerName string `form:"containername"`
}
