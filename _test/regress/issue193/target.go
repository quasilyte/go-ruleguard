package main

import (
	"fmt"

	"github.com/go-ruleguard/rg1"
	"go.uber.org/zap"
)

type Foo struct {
}

type Bar struct {
}

func (f *Foo) String() string {
	return "foo"
}

func (f *Bar) Run() {
	fmt.Printf("Running\n")
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	f := &Foo{}
	sugar.Infof("Stringer %+v", f)

	var err error
	sugar.Info(err)
	sugar.Info(err, "")

	var w rg1.Worker = &Bar{}
	w.Run()

	sugar.Info("Test")
}
