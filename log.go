package main

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"sync"
)

var (
	singleton *zap.Logger
	once      sync.Once
)

func InitLogger(logFile string) error {
	var err error
	once.Do(func() {
		singleton, err = newLogger(logFile)
	})
	return err
}

func GetLogger() *zap.Logger {
	if singleton == nil {
		panic("logger is not initialized")
	}
	return singleton
}

func newLogger(logFile string) (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	
	writeSyncer := zapcore.AddSync(os.Stdout)
	if logFile != "" {
		lumberJackLogger := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
			Compress:   true,
		}
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lumberJackLogger))
	}
	
	encoder := zapcore.NewJSONEncoder(config.EncoderConfig)
	level := zap.InfoLevel
	
	core := zapcore.NewCore(encoder, writeSyncer, level)
	logger := zap.New(core, zap.AddCaller())
	
	return logger, nil
}
