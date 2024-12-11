package internal

import (
	"io"
	"os"
	"time"

	"github.com/Omashu-Data/api-data/riot-petitions/pkg/core"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func SetupLogger(stderr *os.File) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	var writers []io.Writer

	// Create a console writer for terminal output
	consoleWriter := zerolog.ConsoleWriter{
		Out:        stderr,
		TimeFormat: time.RFC3339,
	}
	writers = append(writers, consoleWriter)

	if core.Config.LOGS_TO_FILE {
		// Ensure the directory exists
		logDir := "/app/logs"
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				panic(err)
			}
		}

		logFile, err := os.OpenFile(
			logDir+"/riot-limiter.log",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0664,
		)
		if err != nil {
			panic(err)
		}
		writers = append(writers, logFile)
	}

	// Combine writers based on the writeToFile flag
	multi := zerolog.MultiLevelWriter(writers...)
	log.Logger = zerolog.New(multi).Level(zerolog.InfoLevel).With().Timestamp().Logger()
}
