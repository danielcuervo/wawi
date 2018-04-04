// Package logx is a logging package inspired by Sirupsen/logrus and
// uber-common/zap that follows these guidelines
// https://socialpoint.atlassian.net/wiki/display/BAC/Logging+guidelines+for+Golang+applications
package logx

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"time"
)

// Field is a key/value pair associated to a log.
type Field struct {
	Key   string
	Value interface{}
}

// F returns a new log field with the provided key and value
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Logging levels
const (
	DebugLevel Level = iota + 1
	InfoLevel
)

// Level type
type Level uint8

func (l Level) String() string {
	switch l {
	case 1:
		return "DEBU"
	case 2:
		return "INFO"
	default:
		return "????"
	}
}

// DefaultMinLevel is the minimum debug level for which the logs will appear.
var DefaultMinLevel = DebugLevel

// DefaultFileSkipLevel is the number of stack frames to ascend to get the calling file
var DefaultFileSkipLevel = 3

// a log entry has a message, some fields (optional) and a log level
type entry struct {
	message string
	fields  []Field
	level   Level
	time    *time.Time
	file    string
}

// Logger defines the log methods Debug and Info
// and also provides a level getter
type Logger interface {
	Debug(string, ...Field)
	Info(string, ...Field)
	Level() Level
}

// A Log implements Logger and has a marshaler, a writer and a minimum log level.
type Log struct {
	marshaler     Marshaler
	writer        io.Writer
	level         Level
	withoutTime   bool
	fileSkipLevel int
}

// Debug logs a message at level Debug
func (l *Log) Debug(message string, fields ...Field) {
	if DebugLevel >= l.level {
		l.log(DebugLevel, message, fields...)
	}
}

// Info logs a message at level Info
func (l *Log) Info(message string, fields ...Field) {
	if InfoLevel >= l.level {
		l.log(InfoLevel, message, fields...)
	}
}

// Level returns the logger level
func (l *Log) Level() Level {
	return l.level
}

func (l *Log) log(level Level, message string, fields ...Field) {
	var t *time.Time
	if !l.withoutTime {
		time := time.Now()
		t = &time
	}
	entry := &entry{
		message: message,
		fields:  fields,
		level:   level,
		time:    t,
		file:    fileInfo(l.fileSkipLevel),
	}
	data, err := l.marshaler.Marshal(entry)
	if err == nil {
		_, _ = l.writer.Write(data)
	}
	// @TODO what to do here? metric?
}

// DefaultWriter is the writer default to all loggers
var DefaultWriter = os.Stdout

// NewLogstash creates a new logstash compatible logger
func NewLogstash(channel, product, application string, opts ...Option) *Log {
	options := &options{}
	for _, opt := range opts {
		opt(options)
	}

	if options.marshaler == nil {
		options.marshaler = NewLogstashMarshaler(channel, product, application)
	}
	if options.writer == nil {
		options.writer = DefaultWriter
	}
	if options.level == 0 {
		options.level = DefaultMinLevel
	}
	if options.fileSkipLevel == 0 {
		options.fileSkipLevel = DefaultFileSkipLevel
	}

	return loggerFromOptions(options)
}

// New creates a basic logger with the default values.
func New(opts ...Option) *Log {
	options := &options{}
	for _, opt := range opts {
		opt(options)
	}

	if options.marshaler == nil {
		options.marshaler = new(HumanMarshaler)
	}
	if options.writer == nil {
		options.writer = DefaultWriter
	}
	if options.level == 0 {
		options.level = DefaultMinLevel
	}
	if options.fileSkipLevel == 0 {
		options.fileSkipLevel = DefaultFileSkipLevel
	}

	return loggerFromOptions(options)
}

// NewDummy creates a logger for testing purposes.
func NewDummy(opts ...Option) *Log {
	options := &options{}
	for _, opt := range opts {
		opt(options)
	}

	if options.marshaler == nil {
		options.marshaler = new(DummyMarshaler)
	}
	if options.writer == nil {
		options.writer = ioutil.Discard
	}
	if options.level == 0 {
		options.level = DefaultMinLevel
	}
	if options.fileSkipLevel == 0 {
		options.fileSkipLevel = DefaultFileSkipLevel
	}

	return loggerFromOptions(options)
}

func loggerFromOptions(opts *options) *Log {
	return &Log{
		opts.marshaler,
		opts.writer,
		opts.level,
		opts.withoutTime,
		opts.fileSkipLevel,
	}
}

func fileInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "<???>"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	return fmt.Sprintf("%s:%d", file, line)
}
