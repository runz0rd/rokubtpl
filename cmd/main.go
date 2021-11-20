package main

import (
	"flag"
	"time"

	"github.com/runz0rd/rokubtpl"
	"github.com/sirupsen/logrus"
)

func main() {
	var configFlag string
	flag.StringVar(&configFlag, "config", "config.yaml", "config file path")
	flag.Parse()

	logger := logrus.New()
	c, err := rokubtpl.LoadConfig(configFlag)
	if err != nil {
		logger.Fatal(err)
	}
	if c.Debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	if err := run(logger, c); err != nil {
		logger.Fatal(err)
	}
}

func run(logger *logrus.Logger, c *rokubtpl.Config) error {
	log := logrus.NewEntry(logger)
	pl := rokubtpl.NewJarPrivateListening(c.PrivateListeningBinPath)
	rbt, err := rokubtpl.New(log, c, pl)
	if err != nil {
		return err
	}
	for {
		log.Debug("checking if up")
		isUp := rbt.IsRokuUp()
		if isUp && !rbt.IsPlStarted() {
			rbt.Start()
		}
		if !isUp && rbt.IsPlStarted() {
			rbt.Stop()
		}
		time.Sleep(time.Duration(c.CheckDelaySec) * time.Second)
	}
}
