package main

import (
	"fmt"
	"log/slog"
)

type SlogAdapter struct {
	Adapter *slog.Logger
}

func (s SlogAdapter) Fatal(args ...interface{}) {
	s.Error(args)
}

func (s SlogAdapter) Fatalf(format string, args ...interface{}) {
	s.Errorf(format, args)
}

func (s SlogAdapter) Fatalln(args ...interface{}) {
	s.Errorln(args)
}

func (s SlogAdapter) Panic(args ...interface{}) {
	s.Error(args)
	panic(fmt.Sprint(args...))
}

func (s SlogAdapter) Panicf(format string, args ...interface{}) {
	s.Errorf(format, args)
	panic(fmt.Sprintf(format, args...))
}

func (s SlogAdapter) Panicln(args ...interface{}) {
	s.Errorln(args)
	panic(fmt.Sprintln(args...))
}

func (s SlogAdapter) Print(args ...interface{}) {
	s.Info(args)
}

func (s SlogAdapter) Printf(format string, args ...interface{}) {
	s.Infof(format, args)
}

func (s SlogAdapter) Println(args ...interface{}) {
	s.Infoln(fmt.Sprintln(args...))
}

func (s SlogAdapter) Debug(args ...interface{}) {
	s.Adapter.Debug(fmt.Sprint(args...))
}

func (s SlogAdapter) Debugf(format string, args ...interface{}) {
	s.Adapter.Debug(format, args)
}

func (s SlogAdapter) Debugln(args ...interface{}) {
	s.Adapter.Debug(fmt.Sprintln(args...))
}

func (s SlogAdapter) Error(args ...interface{}) {
	s.Adapter.Error(fmt.Sprint(args...))
}

func (s SlogAdapter) Errorf(format string, args ...interface{}) {
	s.Adapter.Error(format, args)
}

func (s SlogAdapter) Errorln(args ...interface{}) {
	s.Adapter.Error(fmt.Sprintln(args...))
}

func (s SlogAdapter) Info(args ...interface{}) {
	s.Adapter.Info(fmt.Sprint(args...))
}

func (s SlogAdapter) Infof(format string, args ...interface{}) {
	s.Adapter.Info(format, args)
}

func (s SlogAdapter) Infoln(args ...interface{}) {
	s.Adapter.Info(fmt.Sprintln(args...))
}

func (s SlogAdapter) Warn(args ...interface{}) {
	s.Adapter.Warn(fmt.Sprint(args...))
}

func (s SlogAdapter) Warnf(format string, args ...interface{}) {
	s.Adapter.Warn(format, args)
}

func (s SlogAdapter) Warnln(args ...interface{}) {
	s.Adapter.Warn(fmt.Sprintln(args...))
}
