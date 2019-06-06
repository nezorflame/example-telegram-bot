package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nezorflame/example-telegram-bot/internal/pkg/config"
	"github.com/nezorflame/example-telegram-bot/internal/pkg/db"
	"github.com/nezorflame/example-telegram-bot/pkg/telegram"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var configName string

func init() {
	// get flags, init logger
	pflag.StringVar(&configName, "config", "config", "Config file name")
	level := pflag.String("log-level", "INFO", "Logrus log level (DEBUG, WARN, etc.)")
	pflag.Parse()

	logLevel, err := log.ParseLevel(*level)
	if err != nil {
		log.WithError(err).Fatal("Unable to parse log level")
	}
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(logLevel)

	if configName == "" {
		pflag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	// init flags and ctx
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// init config and tracing
	log.Info("Starting the bot")
	cfg, err := config.New(configName)
	if err != nil {
		log.WithError(err).Fatal("Unable to parse config")
	}
	log.Info("Config parsed")

	// init DB
	dbInstance, err := db.New(cfg.GetString("db.path"), cfg.GetDuration("db.timeout"))
	if err != nil {
		log.WithError(err).Fatal("Unable to init DB")
	}
	log.Info("DB initiated")

	// create bot
	bot, err := telegram.NewBot(ctx, cfg)
	if err != nil {
		log.WithError(err).Fatal("Unable to create bot")
	}
	log.Info("Bot created")

	// init graceful stop chan
	log.Debug("Initiating system signal watcher")
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	go func() {
		sig := <-gracefulStop
		log.Warnf("Caught sig %+v, stopping the app", sig)
		cancel()
		bot.Stop()
		dbInstance.Close(false)
		time.Sleep(200 * time.Millisecond)
		os.Exit(0)
	}()

	// start the bot
	log.Info("Starting the bot")
	bot.Start()
}
