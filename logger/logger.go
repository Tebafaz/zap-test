package logger

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LogFormatJson    = "json"
	LogFormatConsole = "console"
)

func NewLogger(projectName, serviceName, branchName, serviceVersion string, conf *Config) (
	_ *zap.Logger,
	err error,
) {
	if conf.Cores == nil || len(conf.Cores) == 0 {
		return zap.New(zapcore.NewNopCore()), nil
	}

	var (
		cores = make([]zapcore.Core, 0)
	)

	for _, coreCfg := range conf.Cores {
		core, err := newCore(coreCfg)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create logger core")
		}

		cores = append(cores, core)
	}

	var options = make([]zap.Option, 0, 8)
	if conf.Caller {
		options = append(options, zap.AddCaller())
	}

	if conf.Development {
		options = append(options, zap.Development())
	}

	var stacktraceLevel zap.AtomicLevel
	for _, core := range conf.Cores {
		if len(core.Stacktrace) > 0 {
			if err = stacktraceLevel.UnmarshalText([]byte(core.Stacktrace)); err != nil {
				return nil, err
			}

			options = append(options, zap.AddStacktrace(stacktraceLevel))
		}
	}

	options = append(options, zap.Fields(
		zap.String("project", projectName),
		zap.String("service", serviceName),
		zap.String("branch", branchName),
		zap.String("version", serviceVersion),
	))

	instance := zap.New(zapcore.NewTee(cores...), options...)

	return instance, nil
}

// newCore Создание нового Кора из конфига
func newCore(coreCfg Core) (core zapcore.Core, err error) {
	var writer zapcore.WriteSyncer
	if coreCfg.OutputConfig.Path != "" {
		if writer, err = addFileWriter(writer, coreCfg.OutputConfig); err != nil {
			return nil, err
		}
	}
	// writer == nil for backward compatibility with old configs
	if coreCfg.OutputConfig.UseStdOut || writer == nil {
		writer = zapcore.Lock(os.Stdout)
	}

	var level zap.AtomicLevel
	if len(coreCfg.Level) > 0 {
		if err = level.UnmarshalText([]byte(coreCfg.Level)); err != nil {
			return nil, err
		}
	}

	encoderCfg := zap.NewProductionEncoderConfig() // Use the same json log format everywhere (for journal beat)

	var priority zap.LevelEnablerFunc = func(lvl zapcore.Level) bool {
		return lvl >= level.Level()
	}

	switch coreCfg.Encoding {
	case LogFormatJson:
		core = zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), writer, priority)
	case LogFormatConsole:
		encoderCfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder
		if coreCfg.OutputConfig.UseCapitalColor {
			encoderCfg.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
				if l == zapcore.DPanicLevel {
					enc.AppendString(fmt.Sprintf("\x1b[%dm%s\x1b[0m", uint8(31), "CRITICAL"))
					return
				}

				zapcore.CapitalColorLevelEncoder(l, enc)
			}
		}

		core = zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), writer, priority)
	default:
		return nil, fmt.Errorf("unknown encoding %s", coreCfg.Encoding)
	}

	return core, nil
}

// addFileWriter добавить писателя в файл к писателю Кора
func addFileWriter(writer zapcore.WriteSyncer, cfg OutputConfig) (zapcore.WriteSyncer, error) {
	var (
		err        error
		fileWriter *FlushTimerBuff
	)
	if fileWriter, err = NewFileWriter(cfg); err != nil {
		return nil, err
	}
	fileWriter.FileFlashWorker()

	if fileWriter.Writer != nil {
		if writer == nil {
			writer = zapcore.Lock(fileWriter)
			return writer, nil
		}
		writer = zap.CombineWriteSyncers(writer, fileWriter)
	}

	return writer, nil
}
