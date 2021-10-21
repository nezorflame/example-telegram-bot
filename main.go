package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nezorflame/example-telegram-bot/internal/pkg/bolt"
	"github.com/nezorflame/example-telegram-bot/internal/pkg/config"
	"github.com/nezorflame/example-telegram-bot/pkg/telegram"

	"github.com/sirupsen/logrus"
)

// Config flags.
var (
	configName  string
	logrusLevel logrus.Level
)

// Init the flags.
func init() {
	flag.StringVar(&configName, "config", "config", "Config file name")
	logLevel := flag.String("log-level", "INFO", "Logrus log level (DEBUG, WARN, etc.)")
	flag.Parse()
	if logLevel == nil {
		flag.PrintDefaults()
		os.Exit(1)
	}

	var err error
	if logrusLevel, err = logrus.ParseLevel(*logLevel); err != nil {
		fmt.Printf("Log level '%s' is incorrect: %s\n", *logLevel, err)
		flag.PrintDefaults()
		os.Exit(1)
	}

	if configName == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	// init flags and ctx
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// init logger
	log := &logrus.Logger{
		Out:       os.Stdout,
		Formatter: &logrus.TextFormatter{FullTimestamp: true},
		Hooks:     make(logrus.LevelHooks),
		Level:     logrusLevel,
	}

	// init config
	cfg, err := config.New(configName)
	if err != nil {
		log.WithError(err).Fatal("Unable to parse config")
	}
	log.Info("Config parsed")

	// init DB
	db, err := bolt.New(cfg.GetString("db.path"), cfg.GetDuration("db.timeout"))
	if err != nil {
		log.WithError(err).Fatal("Unable to init DB")
	}
	log.Info("DB initiated")
	defer db.Close(false)

	// create bot
	bot, err := telegram.NewBot(cfg, log)
	if err != nil {
		log.WithError(err).Fatal("Unable to create bot")
	}
	log.Info("Bot created")

	// init graceful stop chan
	log.Debug("Initiating system signal watcher")
	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	// start the bot
	log.Info("Starting the bot")
	bot.Start(ctx)
	log.Info("Started the bot, listening to the updates...")
	defer bot.Stop()

	// watch context and syscalls
	select {
	case sig := <-gracefulStop:
		log.Warnf("Caught sig %+v, stopping the app", sig)
		cancel()
		return
	case <-ctx.Done():
		log.Warnf("Context closed (%s), exiting application", ctx.Err())
		return
	}
}
