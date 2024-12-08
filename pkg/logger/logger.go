package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"os"
	"time"
)

const serviceName = "investment-game-backend"

type Config struct {
	Level         string
	FilePath      string
	NeedLogToFile bool
}

func New(cfg Config) *zerolog.Logger {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		panic(err)
	}

	output := zerolog.ConsoleWriter{
		TimeFormat: time.RFC3339Nano,
		Out:        os.Stdout,
	}

	writers := []io.Writer{output}

	if cfg.NeedLogToFile {
		file, err := openOrCreateLogFile(cfg.FilePath)
		if err != nil {
			panic(err)
		}
		writers = append(writers, file)
	}

	multi := zerolog.MultiLevelWriter(writers...)

	l := zerolog.New(multi).With().Caller().Timestamp().Logger()
	l.Level(level)

	return &l
}

func openOrCreateLogFile(filepath string) (*os.File, error) {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		if err = os.Mkdir(filepath, 0777); err != nil {
			return nil, fmt.Errorf("create log directory: %w", err)
		}
	}
	logsFilePath := fmt.Sprintf("%s/%s.log", filepath, serviceName)
	file, err := os.OpenFile(logsFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return nil, fmt.Errorf("open or create file: %w", err)
	}
	return file, nil
}
