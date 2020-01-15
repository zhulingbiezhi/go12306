package logger

import (
	"bytes"
	"chat/utils/gls"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/jinzhu/gorm"
	"os"
	"runtime"
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
var LogManager = map[string]*log.Logger{}
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
	levelStr := fmt.Sprintf("[%s] ", entry.Level.String())
	buf.WriteString(fileStr)
	buf.WriteString(reqIDStr)
	buf.WriteString("\n")
	buf.WriteString(timeStr)
	buf.WriteString(levelStr)
	delete(data, FieldKeyRequestID)
	delete(data, FieldKeyFileName)
	buf.WriteString(fmt.Sprintf("[ %s ]", entry.Message))
	buf.WriteString("\n")
	if len(data) > 0 {
		buf.WriteString(awsutil.Prettify(data))
	}
	buf.WriteString("\n")
	return buf.Bytes(), nil
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
	loggerLevel = log.ErrorLevel
	defaultLog = &log.Logger{
		Out:       os.Stdout,
		Formatter: &LogFormatter{},
		Hooks:     make(log.LevelHooks),
		Level:     loggerLevel,
	}
}

type GormLogger struct {
	logInstance *log.Entry
}

func NewGormLogger() *GormLogger {
	return &GormLogger{logInstance: GetLogger()}
}

func (g *GormLogger) Print(v ...interface{}) {
	g.logInstance.WithFields(log.Fields{"module": "gorm", "type": "sql"}).Print(gorm.LogFormatter(v...)...)
}
