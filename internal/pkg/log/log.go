package log

import (
	"context"

	"github.com/sirupsen/logrus"
)

type TraceIDType string

const MSG_ID TraceIDType = "MsgID"

var logger = logrus.New()

func init() {
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}

func getTraceID(ctx context.Context) string {
	if ctx != nil {
		if traceID, ok := ctx.Value(MSG_ID).(string); ok {
			return traceID
		}
	}
	return "DEFAULT-0000"
}

func SetLevel(level logrus.Level) {
	logger.SetLevel(level)
}

func Painc(ctx context.Context, msg string) {
	traceID := getTraceID(ctx)
	logger.Panicf("%s [%s]", msg, traceID)
}

func Error(ctx context.Context, msg string) {
	traceID := getTraceID(ctx)
	logger.Errorf("%s [%s]", msg, traceID)
}

func Warn(ctx context.Context, msg string) {
	traceID := getTraceID(ctx)
	logger.Warnf("%s [%s]", msg, traceID)
}

func Info(ctx context.Context, msg string) {
	traceID := getTraceID(ctx)
	logger.Infof("%s [%s]", msg, traceID)
}

func Debug(ctx context.Context, msg string) {
	traceID := getTraceID(ctx)
	logger.Debugf("%s [%s]", msg, traceID)
}
