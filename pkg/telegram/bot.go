package telegram

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Bot describes Telegram bot
type Bot struct {
	ctx context.Context
	api *tgbotapi.BotAPI
	cfg *viper.Viper
}

// NewBot creates new instance of Bot
func NewBot(ctx context.Context, cfg *viper.Viper) (*Bot, error) {
	if cfg == nil {
		return nil, errors.New("empty config")
	}

	api, err := tgbotapi.NewBotAPI(cfg.GetString("telegram.token"))
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect to Telegram")
	}
	_ = tgbotapi.SetLogger(log.WithField("source", "telegram-api"))
	if cfg.GetBool("telegram.debug") {
		log.Debug("Enabling debug mode for bot")
		api.Debug = true
	}

	log.Debugf("Authorized on account %s", api.Self.UserName)
	return &Bot{api: api, cfg: cfg, ctx: ctx}, nil
}

// Start starts to listen the bot updates channel
func (b *Bot) Start() {
	update := tgbotapi.NewUpdate(0)
	update.Timeout = b.cfg.GetInt("telegram.timeout")
	updates, err := b.api.GetUpdatesChan(update)
	if err != nil {
		log.WithError(err).Fatal("Unable to start listening to bot updates")
	}
	b.listen(updates)
}

// Stop stops the bot
func (b *Bot) Stop() {
	b.api.StopReceivingUpdates()
}

func (b *Bot) listen(updates tgbotapi.UpdatesChannel) {
	for u := range updates {
		if u.Message == nil { // ignore any non-Message Updates
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
