package errors

import "fmt"

func Errorf(err error, format string, v ...interface{}) error {
	v = append(v, fmt.Sprint(err))
	return fmt.Errorf(fmt.Sprintf(format+": [ %s ]", v...))
}
