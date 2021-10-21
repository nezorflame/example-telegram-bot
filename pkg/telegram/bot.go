package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Bot describes Telegram bot
type Bot struct {
	api *tgbotapi.BotAPI
	cfg *viper.Viper
	log *logrus.Logger
}

// NewBot creates new instance of Bot
func NewBot(cfg *viper.Viper, log *logrus.Logger) (*Bot, error) {
	if cfg == nil {
		return nil, errors.New("empty config")
	}

	api, err := tgbotapi.NewBotAPI(cfg.GetString("telegram.token"))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Telegram: %w", err)
	}
	_ = tgbotapi.SetLogger(log.WithField("source", "telegram-api"))
	if cfg.GetBool("telegram.debug") {
		log.Debug("Enabling debug mode for bot")
		api.Debug = true
	}

	log.Debugf("Authorized on account %s", api.Self.UserName)
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
			b.log.Warning("Context closed - stopping listening to the updates: %w", ctx.Err())
			return
		case u := <-updates:
			if u.Message == nil { // ignore any non-Message updates
				continue
			}

			switch {
			case strings.HasPrefix(u.Message.Text, b.cfg.GetString("commands.start")):
				go b.hello(u.Message)
			case strings.HasPrefix(u.Message.Text, b.cfg.GetString("commands.help")):
				log.WithField("user_id", u.Message.From.ID).Debug("Got help request")
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
	log.WithField("chat_id", chatID).WithField("msg_id", msgID).Debug("Sending reply")
	msg := tgbotapi.NewMessage(chatID, fmt.Sprint(text))
	if msgID != 0 {
		msg.ReplyToMessageID = msgID
	}
	msg.ParseMode = tgbotapi.ModeMarkdown

	if _, err := b.api.Send(msg); err != nil {
		log.Errorf("Unable to send the message: %v", err)
		return
	}
}
