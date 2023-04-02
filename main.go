package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/exp/slog"

	"github.com/nezorflame/example-telegram-bot/internal/bolt"
	"github.com/nezorflame/example-telegram-bot/internal/bot"
	"github.com/nezorflame/example-telegram-bot/internal/config"
)

// Config flags.
var (
	configName string
	slogLevel  slog.Level
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

	if err := slogLevel.UnmarshalText([]byte(*logLevel)); err != nil {
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
	slogOptions := slog.HandlerOptions{
		AddSource: true,
		Level:     slogLevel,
	}
	log := slog.New(slogOptions.NewTextHandler(os.Stdout))
	log.Info("Launching the bot...")

	// error reporting
	var err error
	defer func() {
		if panicErr := recover(); panicErr != nil {
			log.Error("Got a panic", "error", panicErr)
			os.Exit(1)
		}

		if err != nil {
			log.Error("Got an error", "error", err)
			os.Exit(1)
		}
	}()

	// init config
	cfg, err := config.New(configName)
	if err != nil {
		err = fmt.Errorf("unable to parse config: %w", err)
		return
	}
	log.Info("Config parsed")

	// init DB
	db, err := bolt.New(
		cfg.GetString("db.path"),
		cfg.GetDuration("db.timeout"),
		log,
	)
	if err != nil {
		err = fmt.Errorf("unable to init DB: %w", err)
		return
	}
	log.Info("DB initiated")
	defer db.Close(false)

	// create tgBot
	tgBot, err := bot.New(cfg, log, slogLevel, db)
	if err != nil {
		err = fmt.Errorf("unable to create bot: %w", err)
		return
	}
	log.Info("Bot created")

	// init graceful stop chan
	log.Debug("Initiating system signal watcher")
	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	// start the bot
	log.Info("Starting the bot")
	tgBot.Start(ctx)
	log.Info("Started the bot, listening to the updates...")
	defer tgBot.Stop()

	// watch context and syscalls
	select {
	case sig := <-gracefulStop:
		log.Warn("Caught a signal, stopping the app", "signal", sig)
		cancel()
		return
	case <-ctx.Done():
		log.Warn("Context closed, exiting application", "error", ctx.Err())
		return
	}
}
