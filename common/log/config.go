package log

import (
	"errors"
)

type ConfFileWriter struct {
	On              bool   `toml:"On"`
	LogPath         string `toml:"LogPath"`
	RotateLogPath   string `toml:"RotateLogPath"`
	WfLogPath       string `toml:"WfLogPath"`
	RotateWfLogPath string `toml:"RotateWfLogPath"`
}

type ConfConsoleWriter struct {
	On    bool `toml:"On"`
	Color bool `toml:"Color"`
}

type LogConfig struct {
	Level string            `toml:"LogLevel"`
	FW    ConfFileWriter    `toml:"FileWriter"`
	CW    ConfConsoleWriter `toml:"ConsoleWriter"`
}

func SetupLogInstanceWithConf(lc LogConfig, logger *Logger) (err error) {
	if lc.FW.On {
		if len(lc.FW.LogPath) > 0 {
			w := NewFileWriter()
			//设置普通日志路径
			w.SetFileName(lc.FW.LogPath)
			//用于设置日志轮转的路径模式，控制日志的分片存储。
			w.SetPathPattern(lc.FW.RotateLogPath)
			w.SetLogLevelFloor(TRACE)
			if len(lc.FW.WfLogPath) > 0 {
				w.SetLogLevelCeil(INFO)
			} else {
				w.SetLogLevelCeil(ERROR)
			}
			logger.Register(w)
		}

		if len(lc.FW.WfLogPath) > 0 {
			wfw := NewFileWriter()
			//设置警告日志路径
			wfw.SetFileName(lc.FW.WfLogPath)
			//用于设置日志轮转的路径模式，控制日志的分片存储。
			wfw.SetPathPattern(lc.FW.RotateWfLogPath)
			wfw.SetLogLevelFloor(WARNING)
			wfw.SetLogLevelCeil(ERROR)
			logger.Register(wfw)
		}
	}

	if lc.CW.On {
		w := NewConsoleWriter()
		w.SetColor(lc.CW.Color)
		logger.Register(w)
	}
	switch lc.Level {
	case "trace":
		logger.SetLevel(TRACE)

	case "debug":
		logger.SetLevel(DEBUG)

	case "info":
		logger.SetLevel(INFO)

	case "warning":
		logger.SetLevel(WARNING)

	case "error":
		logger.SetLevel(ERROR)

	case "fatal":
		logger.SetLevel(FATAL)

	default:
		err = errors.New("Invalid log level")
	}
	return
}

func SetupDefaultLogWithConf(lc LogConfig) (err error) {
	//创建一个新的日志实例，并启动写入协程
	defaultLoggerInit()
	return SetupLogInstanceWithConf(lc, logger_default)
}
