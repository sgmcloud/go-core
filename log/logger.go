package log

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
}

type Field struct {
	Key   string
	Value interface{}
}

func Any(key string, value interface{}) Field {
	return Field{
		Key:   key,
		Value: value,
	}
}

func Error(err error) Field {
	return Field{
		Key:   "error",
		Value: err,
	}
}
