package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nezorflame/example-telegram-bot/internal/bolt"
	"github.com/nezorflame/example-telegram-bot/internal/bot"
	"github.com/nezorflame/example-telegram-bot/internal/config"
)

// Config flags.
var (
	envFile   string
	slogLevel slog.Level
)

// Init the flags.
func init() {
	var err error
	defer func() {
		if err != nil {
			flag.Usage()
			os.Exit(1)
		}
	}()

	flag.StringVar(&envFile, "env", "", "Config file name")
	logLevel := flag.String("log", "INFO", "Log level (DEBUG, WARN, etc.)")
	flag.Parse()

	// validate log level, if set
	if err = slogLevel.UnmarshalText([]byte(*logLevel)); err != nil {
		err = fmt.Errorf("unable to marshal log level: %w", err)
		return
	}

	// validate dotenv file, if set
	if envFile != "" {
		if _, err = os.Stat(envFile); err != nil {
			err = fmt.Errorf("unable to read dotenv file: %w", err)
			return
		}
	}
}

func main() {
	// init flags and ctx
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// init logger
	slogOpts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slogLevel,
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, slogOpts))
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
	cfg, err := config.New(envFile)
	if err != nil {
		err = fmt.Errorf("unable to parse config: %w", err)
		return
	}
	log.Info("Config parsed")

	// init DB
	db, err := bolt.New(cfg.DBPath, cfg.DBTimeout, log)
	if err != nil {
		err = fmt.Errorf("unable to init DB: %w", err)
		return
	}
	log.Info("DB initiated")
	defer db.Close(false)

	// create Telegram bot
	tgBot, err := bot.New(cfg, log, db)
	if err != nil {
		err = fmt.Errorf("unable to create bot: %w", err)
		return
	}
	log.Info("Bot created")

	// init graceful stop chan
	log.Debug("Initiating system signal watcher")
	graceful := make(chan os.Signal, 1)
	signal.Notify(graceful, syscall.SIGTERM)
	signal.Notify(graceful, syscall.SIGINT)

	// start the bot
	log.Info("Starting the bot")
	tgBot.Start()
	log.Info("Started the bot, listening to the updates...")
	defer tgBot.Stop()

	// watch context and syscalls
	select {
	case sig := <-graceful:
		log.Warn("Caught a signal, stopping the app", "signal", sig)
		cancel()
		return
	case <-ctx.Done():
		log.Warn("Context closed, exiting application", "error", ctx.Err())
		return
	}
}
