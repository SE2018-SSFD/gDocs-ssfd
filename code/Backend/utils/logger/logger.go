package logger

import (
	"github.com/kataras/golog"
)

var logger *golog.Logger

var (
	Debug	=	logger.Debug
	Debugf	=	logger.Debugf

	Info	=	logger.Info
	Infof	=	logger.Infof

	Warn	=	logger.Warn
	Warnf	=	logger.Warnf

	Error	=	logger.Error
	Errorf	=	logger.Errorf

	Fatal	=	logger.Fatal
	Fatalf	=	logger.Fatalf
)

func SetLogger(l *golog.Logger) {
	logger = l

	Debug	=	logger.Debug
	Debugf	=	logger.Debugf

	Info	=	logger.Info
	Infof	=	logger.Infof

	Warn	=	logger.Warn
	Warnf	=	logger.Warnf

	Error	=	logger.Error
	Errorf	=	logger.Errorf

	Fatal	=	logger.Fatal
	Fatalf	=	logger.Fatalf
}

func SetLevel(level string) {
	logger.SetLevel(level)
}