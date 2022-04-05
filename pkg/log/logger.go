package log

import "github.com/sirupsen/logrus"

const (
	errLabel = "_err"
)

// NewLogger create project-level logger instance.
func NewLogger() *baseLogger { // nolint:revive,golint
	logrus.SetReportCaller(true)

	formatter := &logrus.JSONFormatter{} // nolint:exhaustivestruct
	logrus.SetFormatter(formatter)

	return &baseLogger{
		label: map[string]interface{}{},
	}
}

var _ Logger = &baseLogger{} // nolint:exhaustivestruct

type baseLogger struct {
	label map[string]interface{}
}

func (l baseLogger) WithErr(err error) Logger { // nolint:ireturn
	l.label[errLabel] = err.Error()

	return l
}

func (l baseLogger) Info(args ...interface{}) {
	logrus.WithFields(l.label).Info(args...)
	l.clear()
}

func (l baseLogger) Infof(format string, args ...interface{}) {
	logrus.WithFields(l.label).Infof(format, args...)
	l.clear()
}

func (l baseLogger) Warn(args ...interface{}) {
	logrus.WithFields(l.label).Warn(args...)
	l.clear()
}

func (l baseLogger) Fatal(args ...interface{}) {
	logrus.WithFields(l.label).Fatal(args...)
	l.clear()
}

func (l baseLogger) clear() {
	for k := range l.label {
		delete(l.label, k)
	}
}
