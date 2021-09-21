package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ddvk/rmfakecloud/internal/app"
	"github.com/ddvk/rmfakecloud/internal/cli"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var version string

func main() {
	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Println("Version: ", version)
		fmt.Printf(`
Commands:
	setuser		create users / reset passwords
	listusers	list available users
`)
		fmt.Println(config.EnvVars())
	}

	flag.Parse()

	cfg := config.FromEnv()

	//cli
	cmd := cli.New(cfg)
	if cmd.Handle(os.Args) {
		return
	}

	fmt.Fprintln(os.Stderr, "run with -h for all available env variables")
	cfg.Verify()

	logger := log.StandardLogger()
	logger.SetFormatter(&log.TextFormatter{})

	var file, err = os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("cannot open log file for writing")
	} else {
		defer file.Close()
		hook := lfshook.NewHook(file, &logrus.TextFormatter{DisableColors: true})
		logger.Hooks.Add(hook)
	}

	if lvl, err := log.ParseLevel(os.Getenv(config.EnvLogLevel)); err == nil {
		fmt.Println("Log level:", lvl)
		logger.SetLevel(lvl)
	}

	log.Println("Version: ", version)
	// configs
	log.Println("Documents will be saved in:", cfg.DataDir)
	log.Println("Url the device should use:", cfg.StorageURL)
	log.Println("Listening on port:", cfg.Port)

	gin.DefaultWriter = logger.Writer()

	a := app.NewApp(cfg)
	go a.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Stopping the service...")
	a.Stop()
	log.Println("Stopped")
}
