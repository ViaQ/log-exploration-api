package logs

import "errors"

func NotFoundError() error {
	return errors.New("Not Found Error")
}
func InvalidTimeStamp() error {
	return errors.New("incorrect time format: Please Enter Time in the following format YYYY-MM-DD'T'HH:mm:ss.SSS[TIMEZONE ex:'Z']")
}
func InvalidLimit() error {
	return errors.New("invalid \"maxlogs\" value, an integer between 0 to 1000 is required")
}
