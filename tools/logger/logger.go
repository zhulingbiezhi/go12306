package logger

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	log "github.com/sirupsen/logrus"
	"github.com/zhulingbiezhi/go12306/tools/gls"
)

const (
	nocolor = 0
	red     = 31
	green   = 32
	yellow  = 33
	blue    = 36
	gray    = 37
)

const (
	FieldKeyLevel          = "level"
	FieldKeyFileName       = "file"
	FieldKeyRequestID      = "request_id"
	FieldKeyMsg            = "msg"
	FieldKeyTime           = "time"
	defaultTimestampFormat = "2006-01-02 15:04:05"
)

var loggerLevel = log.InfoLevel
var defaultLog *log.Logger

func getFileName() string {
	fileName := ""
	_, file, line, ok := runtime.Caller(2)
	if ok {
		fileName = fmt.Sprintf("%s:%d", file, line)
	}
	return fileName
}

func Debug(v ...interface{}) {
	GetLogger().WithField("file", getFileName()).Debug(v...)
}

func Debugf(format string, v ...interface{}) {
	GetLogger().WithField("file", getFileName()).Debugf(format, v...)
}

func Info(v ...interface{}) {
	GetLogger().WithField("file", getFileName()).Info(v...)
}

func Infof(format string, v ...interface{}) {
	GetLogger().WithField("file", getFileName()).Infof(format, v...)
}

func Warn(v ...interface{}) {
	GetLogger().WithField("file", getFileName()).Warn(v...)
}

func Warnf(format string, v ...interface{}) {
	GetLogger().WithField("file", getFileName()).Warnf(format, v...)
}

func Error(v ...interface{}) {
	GetLogger().WithField("file", getFileName()).Error(v...)
}

func Errorf(format string, v ...interface{}) {
	GetLogger().WithField("file", getFileName()).Errorf(format, v...)
}

func NewErrorf(format string, v ...interface{}) error {
	logs := fmt.Sprintf(format, v...)
	Error(logs)
	return errors.New(logs)
}

func WithFields(fields log.Fields) *log.Entry {
	logger := log.WithFields(fields)
	logger.Logger = &log.Logger{
		Out:       os.Stdout,
		Formatter: &LogFormatter{},
		Hooks:     make(log.LevelHooks),
		Level:     loggerLevel,
	}
	return logger
}

type LogFormatter struct {
	// TimestampFormat sets the format used for marshaling timestamps.
	TimestampFormat string
	FieldMap        map[string]interface{}
}

func (f *LogFormatter) Format(entry *log.Entry) ([]byte, error) {
	data := make(map[string]interface{}, len(entry.Data))
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = defaultTimestampFormat
	}
	buf := bytes.NewBuffer(nil)
	fileStr := fmt.Sprintf("(%s) ", data[FieldKeyFileName])
	timeStr := fmt.Sprintf("[%s] ", entry.Time.Format(timestampFormat))
	reqIDStr := fmt.Sprintf("[%s] ", data[FieldKeyRequestID])
	levelStr := fmt.Sprintf("[%s] ", strings.ToUpper(entry.Level.String())[:1])

	c := GetLevelColor(entry.Level)
	_, _ = fmt.Fprintf(buf, "\x1b[%dm", c)
	buf.WriteString(fileStr)
	buf.WriteString(reqIDStr)
	buf.WriteString("\n")
	buf.WriteString(timeStr)
	buf.WriteString(levelStr)
	delete(data, FieldKeyRequestID)
	delete(data, FieldKeyFileName)
	buf.WriteString(fmt.Sprintf("[ %s ]", entry.Message))
	buf.WriteString("\n")
	buf.WriteString("\x1b[0m")
	if len(data) > 0 {
		buf.WriteString(awsutil.Prettify(data))
	}
	buf.WriteString("\n")
	return buf.Bytes(), nil
}

func GetLevelColor(level log.Level) int {
	switch level {
	case log.DebugLevel:
		return gray
	case log.WarnLevel:
		return yellow
	case log.ErrorLevel, log.FatalLevel, log.PanicLevel:
		return red
	default:
		return blue
	}
}

func GetLogger() *log.Entry {
	data, ok := gls.Get("_log")
	if ok {
		return data.(*log.Entry)
	}
	return log.NewEntry(defaultLog)
}

func GoroutineInit(l *log.Entry) {
	l.Level = loggerLevel
	gls.Set("_log", l)
}

func GoroutineShutdown() {
	gls.Shutdown()
}

func init() {
	defaultLog = &log.Logger{
		Out:       os.Stdout,
		Formatter: &LogFormatter{},
		Hooks:     make(log.LevelHooks),
		Level:     loggerLevel,
	}
}
