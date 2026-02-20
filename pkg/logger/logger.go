package logger

import (
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}

// For now I dont delete this logger.go file ( i switched to slog )
// maybe in the future if i want to write "helper" functions
// (for example, a function that masks passwords in logs or adds specific env tags automatically)

func New(env string) *Logger {
	flags := log.Ldate | log.Ltime | log.LUTC

	if env == "local" {
		flags = log.Ldate | log.Ltime | log.Lshortfile
	}

	return &Logger{
		Logger: log.New(os.Stdout, "", flags),
	}
}
