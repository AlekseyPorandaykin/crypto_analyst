package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level       string   `mapstructure:"level"`
	OutputPaths []string `mapstructure:"output_paths"`
	Stacktrace  bool     `mapstructure:"stacktrace"`
}

var DefaultConf = Config{Level: "DEBUG", OutputPaths: []string{"stdout"}}

func Create(conf Config, opts ...zap.Option) (*zap.Logger, error) {
	level, err := zap.ParseAtomicLevel(conf.Level)
	if err != nil {
		return nil, err
	}
	l, err := zap.Config{
		Level:             level,
		Development:       false,
		Encoding:          "json",
		DisableStacktrace: !conf.Stacktrace,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:     "ts",
			LevelKey:    "level",
			MessageKey:  "message",
			LineEnding:  zapcore.DefaultLineEnding,
			EncodeLevel: zapcore.LowercaseLevelEncoder,
			EncodeTime:  zapcore.RFC3339TimeEncoder,
		},
		OutputPaths: conf.OutputPaths,
	}.Build(opts...)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func CreateGlobal(conf Config, opts ...zap.Option) {
	zap.ReplaceGlobals(zap.Must(Create(conf, opts...)))
}

func InitDefaultLogger() {
	CreateGlobal(DefaultConf)
}
