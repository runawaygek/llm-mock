package internal

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitLogger(lvl zapcore.Level) {
	Logger = NewLogger(lvl)
}

func NewLogger(lvl zapcore.Level) *zap.Logger {
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "time"
	encCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(fmt.Sprintf("%s,%03d",
			t.Format("2006-01-02 15:04:05"),
			t.Nanosecond()/1e6))
	}
	encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encCfg.EncodeCaller = zapcore.ShortCallerEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encCfg),
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(lvl),
	)

	return zap.New(core, zap.AddCaller())
}
