package impl

import "cnb.cool/mliev/open/dwz-server/internal/interfaces"

type LoggerField struct {
	Key   string
	Value any
}

func NewLoggerField(key string, value any) interfaces.LoggerFieldInterface {
	return &LoggerField{
		Key:   key,
		Value: value,
	}
}

func (receiver *LoggerField) GetKey() string {
	return receiver.Key
}

func (receiver *LoggerField) GetValue() any {
	return receiver.Value
}
