package logs

import "errors"

func NotFoundError() error {
	return errors.New("Not Found Error")

}
