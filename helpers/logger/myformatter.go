package logger

import (
	"bytes"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

const (
	nocolor = 0
	red     = 31
	green   = 32
	yellow  = 33
	blue    = 36
	gray    = 37
)

var (
	baseTimestamp time.Time
)

func init() {
	baseTimestamp = time.Now()
}

type myFormatter struct {
	// Set to true to bypass checking for a TTY before outputting colors.
	ForceColors bool

	// Force disabling colors.
	DisableColors bool

	// Disable timestamp logging. useful when output is redirected to logging
	// system that already adds timestamps.
	DisableTimestamp bool

	// Enable logging the full timestamp when a TTY is attached instead of just
	// the time passed since beginning of execution.
	FullTimestamp bool

	// TimestampFormat to use for display when a full timestamp is printed
	TimestampFormat string

	// The fields are sorted by default for a consistent output. For applications
	// that log extremely frequently and don't use the JSON formatter this may not
	// be desired.
	DisableSorting bool

	// QuoteEmptyFields will wrap empty fields in quotes if true
	QuoteEmptyFields bool

	// Whether the logger's out is to a terminal
	isTerminal bool

	sync.Once
}

func (f *myFormatter) init(entry *logrus.Entry) {
}

func (f *myFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}
	if !f.DisableSorting {
		sort.Strings(keys)
	}

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	f.Do(func() { f.init(entry) })
	//isColored := (f.ForceColors || f.isTerminal) && !f.DisableColors
	f.appendKeyValue(b, "level", entry.Level.String())
	if entry.Level == logrus.DebugLevel {
		timestampFormat := f.TimestampFormat
		if timestampFormat == "" {
			timestampFormat = "2006-01-02 15:04:05"
		}
		if !f.DisableTimestamp {
			f.appendKeyValue(b, "time", entry.Time.Format(timestampFormat))
		}

		_, filename, lineno, _ := runtime.Caller(6)

		f.appendKeyValue(b, "p", fmt.Sprintf("%s:%d", CutFilename(filename), lineno))
	}

	if entry.Message != "" {
		f.appendKeyValue(b, "msg", entry.Message)
	}
	for _, key := range keys {
		f.appendKeyValue(b, key, entry.Data[key])
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func CutFilename(filename string) string {
	targetfilename := "test"
	i, prei := 0, 0
	for ; i < len(filename); i++ {
		prei = i
		for ; i < len(filename); i++ {
			if filename[i] == '/' {
				break
			}
		}
		if filename[prei:i] == targetfilename {
			return filename[i:]
		}
	}
	if filename[prei:i] != targetfilename {
		return filename
	}
	return filename[i:]
}

func (f *myFormatter) printColored(b *bytes.Buffer, entry *logrus.Entry, keys []string, timestampFormat string) {
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = gray
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = red
	default:
		levelColor = blue
	}

	levelText := strings.ToUpper(entry.Level.String())[0:4]

	if f.DisableTimestamp {
		fmt.Fprintf(b, "\x1b[%dm%s\x1b[0m %-44s ", levelColor, levelText, entry.Message)
	} else if !f.FullTimestamp {
		fmt.Fprintf(b, "\x1b[%dm%s\x1b[0m[%04d] %-44s ", levelColor, levelText, int(entry.Time.Sub(baseTimestamp)/time.Second), entry.Message)
	} else {
		fmt.Fprintf(b, "\x1b[%dm%s\x1b[0m[%s] %-44s ", levelColor, levelText, entry.Time.Format(timestampFormat), entry.Message)
	}
	for _, k := range keys {
		v := entry.Data[k]
		fmt.Fprintf(b, " \x1b[%dm%s\x1b[0m=", levelColor, k)
		f.appendValue(b, v)
	}
}

func (f *myFormatter) needsQuoting(text string) bool {
	if f.QuoteEmptyFields && len(text) == 0 {
		return true
	}
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.' || ch == '_' || ch == '/' || ch == '@' || ch == '^' || ch == '+') {
			return true
		}
	}
	return false
}

func (f *myFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {

	b.WriteString(key)
	b.WriteByte('=')
	f.appendValue(b, value)
	b.WriteByte(' ')
}

func (f *myFormatter) appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}

	if !f.needsQuoting(stringVal) {
		b.WriteString(stringVal)
	} else {
		b.WriteString(fmt.Sprintf("%q", stringVal))
	}
}
