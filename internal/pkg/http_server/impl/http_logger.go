package impl

import (
	"cnb.cool/mliev/open/dwz-server/internal/interfaces"
)

type HttpLogger struct {
	logger  interfaces.LoggerInterface
	traceId string
}

func NewHttpLogger(logger interfaces.LoggerInterface, traceId string) interfaces.LoggerInterface {
	l := &HttpLogger{
		logger:  logger,
		traceId: traceId,
	}
	return l
}

func (receiver *HttpLogger) Debug(format string, args ...interfaces.LoggerFieldInterface) {
	args = append(args, NewLoggerField("traceId", receiver.traceId))
	receiver.logger.Debug(format, args...)
}

func (receiver *HttpLogger) Info(format string, args ...interfaces.LoggerFieldInterface) {
	args = append(args, NewLoggerField("traceId", receiver.traceId))
	receiver.logger.Info(format, args...)
}

func (receiver *HttpLogger) Notice(format string, args ...interfaces.LoggerFieldInterface) {
	args = append(args, NewLoggerField("traceId", receiver.traceId))
	receiver.logger.Info(format, args...)
}

func (receiver *HttpLogger) Error(format string, args ...interfaces.LoggerFieldInterface) {
	args = append(args, NewLoggerField("traceId", receiver.traceId))
	receiver.logger.Error(format, args...)
}

func (receiver *HttpLogger) Warn(format string, args ...interfaces.LoggerFieldInterface) {
	args = append(args, NewLoggerField("traceId", receiver.traceId))
	receiver.logger.Warn(format, args...)
}

func (receiver *HttpLogger) Fatal(format string, args ...interfaces.LoggerFieldInterface) {
	args = append(args, NewLoggerField("traceId", receiver.traceId))
	receiver.logger.Fatal(format, args...)
}
