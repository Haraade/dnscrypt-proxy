package main

import (
	"fmt"
	"io"
	"os"

	"github.com/jedisct1/dlog"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Logger(logMaxSize, logMaxAge, logMaxBackups int, fileName string) io.Writer {
	if fileName == "/dev/stdout" {
		return os.Stdout
	}

	fileInfo, err := os.Stat(fileName)
	if err == nil {
		if fileInfo.IsDir() {
			dlog.Fatalf("Logger-feil: [%v] er en katalog, ikke en fil", fileName)
		}
	} else if !os.IsNotExist(err) {
		dlog.Fatalf("Logger-feil: Kunne ikke sjekke [%v]: %v", fileName, err)
	}

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
	if err != nil {
		dlog.Fatalf("Logger-feil: Kunne ikke opprette eller åpne [%v]: %v", fileName, err)
	}
	file.Close()

	logger := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    logMaxSize,
		MaxAge:     logMaxAge,
		MaxBackups: logMaxBackups,
		Compress:   true,
		LocalTime:  true,
	}

	fmt.Printf("Logger opprettet: %s (MaxSize: %dMB, MaxAge: %d dager, MaxBackups: %d)\n",
		fileName, logMaxSize, logMaxAge, logMaxBackups)

	return logger
}
