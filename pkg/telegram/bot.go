package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

// Bot describes Telegram bot
type Bot struct {
	api *tgbotapi.BotAPI
	cfg *viper.Viper
	log *slog.Logger
}

// NewBot creates new instance of Bot
func NewBot(cfg *viper.Viper, log *slog.Logger, logLevel slog.Level) (*Bot, error) {
	if cfg == nil {
		return nil, errors.New("empty config")
	}

	api, err := tgbotapi.NewBotAPI(cfg.GetString("telegram.token"))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Telegram: %w", err)
	}

	_ = tgbotapi.SetLogger(slog.NewLogLogger(log.With("source", "telegram-api").Handler(), logLevel))
	if cfg.GetBool("telegram.debug") {
		log.Debug("Enabling debug mode for bot")
		api.Debug = true
	}

	log.Debug("Authorized successfully", "account", api.Self.UserName)
	return &Bot{api: api, cfg: cfg, log: log}, nil
}

// Start starts to listen the bot updates channel
func (b *Bot) Start(ctx context.Context) {
	update := tgbotapi.NewUpdate(0)
	update.Timeout = b.cfg.GetInt("telegram.timeout")
	updates := b.api.GetUpdatesChan(update)
	go b.listen(ctx, updates)
}

// Stop stops the bot
func (b *Bot) Stop() {
	b.api.StopReceivingUpdates()
}

func (b *Bot) listen(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	for {
		select {
		case <-ctx.Done():
			b.log.Warn("Context closed - stopping listening to the updates", "error", ctx.Err())
			return
		case u := <-updates:
			if u.Message == nil { // ignore any non-Message updates
				continue
			}

			switch {
			case strings.HasPrefix(u.Message.Text, b.cfg.GetString("commands.start")):
				go b.hello(u.Message)
			case strings.HasPrefix(u.Message.Text, b.cfg.GetString("commands.help")):
				slog.With("user_id", u.Message.From.ID).Debug("Got help request")
				go b.help(u.Message)
				// case strings.HasPrefix(u.Message.Text, b.cfg.GetString("commands.your_command")):
				// go b.yourBotAction(u.Message)
			}
		}
	}
}

func (b *Bot) hello(msg *tgbotapi.Message) {
	b.reply(msg.Chat.ID, msg.MessageID, b.cfg.GetString("messages.hello"))
}

func (b *Bot) help(msg *tgbotapi.Message) {
	b.reply(msg.Chat.ID, msg.MessageID, b.cfg.GetString("messages.help"))
}

func (b *Bot) reply(chatID int64, msgID int, text string) {
	slog.With("chat_id", chatID).With("msg_id", msgID).Debug("Sending reply")
	msg := tgbotapi.NewMessage(chatID, fmt.Sprint(text))
	if msgID != 0 {
		msg.ReplyToMessageID = msgID
	}
	msg.ParseMode = tgbotapi.ModeMarkdown

	if _, err := b.api.Send(msg); err != nil {
		slog.Error("Unable to send the message", "error", err)
		return
	}
}
