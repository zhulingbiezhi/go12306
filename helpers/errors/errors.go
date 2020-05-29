package errors

import (
	"errors"
	"fmt"
	"github.com/zhulingbiezhi/go12306/helpers/logger"
	"runtime"
)

func Errorf(err error, format string, v ...interface{}) error {
	e := &MyErr{}
	e.fileInfo = getFileName()
	v = append(v, err)
	e.err = fmt.Errorf(format+": %w", v...)
	return e
}

type MyErr struct {
	err      error
	fileInfo string
	msg      string
}

func getFileName() string {
	fileName := ""
	_, file, line, ok := runtime.Caller(2)
	if ok {
		fileName = fmt.Sprintf("%s:%d", file, line)
	}
	return fileName
}

func (e *MyErr) Error() string {
	return fmt.Sprintf("\n[ %s ] %s", e.fileInfo, e.err)
}

func (e *MyErr) Unwrap() error {
	logger.Info("Unwrap:", e.err.Error())
	return e.err
}

func Is(err, errDst error) bool {
	return errors.Is(err, errDst)
}
