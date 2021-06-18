package elastic

import (
	"github.com/ViaQ/log-exploration-api/pkg/logs"
	"strconv"
	"time"
)

func validateParams(params logs.Parameters) error {

	var err error
	if len(params.StartTime) > 0 && len(params.FinishTime) > 0 {

		_, err = time.Parse(time.RFC3339Nano, params.StartTime)
		if err != nil {
			return logs.InvalidTimeStamp()
		}
		_, err = time.Parse(time.RFC3339Nano, params.FinishTime)
		if err != nil {
			return logs.InvalidTimeStamp()
		}
	}
	if len(params.MaxLogs) > 0 {
		maxLogs, err := strconv.Atoi(params.MaxLogs)
		if err != nil || maxLogs < 0 {
			return logs.InvalidLimit()
		}
	}

	return nil
}
