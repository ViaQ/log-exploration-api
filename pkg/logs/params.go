package logs
type Parameters struct {
	Namespace string `json:"namespace"`
	Index string `json:"index"`
	Podname string `json:"podname"`
	StartTime string `json:"starttime"`
	FinishTime string `json:"finishtime"`
	Level string `json:"level"`
}