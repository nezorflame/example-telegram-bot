package bot

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/nezorflame/example-telegram-bot/internal/bolt"
	"github.com/nezorflame/example-telegram-bot/internal/config"
)

type bot struct {
	tg  *tele.Bot
	cfg *config.Config
	db  *bolt.DB
	log *slog.Logger
}

// New creates new instance of Bot
func New(cfg *config.Config, log *slog.Logger, db *bolt.DB) (*bot, error) {
	// validate
	if cfg == nil {
		return nil, errors.New("empty config")
	}
	if log == nil {
		return nil, errors.New("empty logger")
	}
	if db == nil {
		return nil, errors.New("empty DB")
	}

	// create bot
	log.Debug("Connecting to Telegram...")
	tg, err := tele.NewBot(tele.Settings{
		Token:  cfg.TelegramToken,
		Poller: &tele.LongPoller{Timeout: cfg.TelegramTimeout},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Telegram: %w", err)
	}
	log.Debug("Connection established")

	b := &bot{tg: tg, cfg: cfg, db: db, log: log}

	// register handlers
	b.tg.Handle(cfg.CmdStart, func(teleCtx tele.Context) error {
		return b.hello(teleCtx)
	})
	b.tg.Handle(cfg.CmdHelp, func(teleCtx tele.Context) error {
		return b.help(teleCtx)
	})
	b.tg.Handle(tele.OnText, func(teleCtx tele.Context) error {
		return b.handle(teleCtx)
	})

	// return the bot
	log.Debug("Authorized successfully", "account", tg.Me.Username)
	return b, nil
}

// Start starts to listen the bot updates channel.
func (b *bot) Start() {
	b.tg.Start()
}

// Stop stops the bot
func (b *bot) Stop() {
	b.tg.Stop()
}

func (b *bot) hello(teleCtx tele.Context) error {
	b.log.With("chat_id", teleCtx.Chat().ID, "msg_id", teleCtx.Message().ID).Debug("Sending hello reply")
	return teleCtx.Send(b.cfg.MsgHello)
}

func (b *bot) help(teleCtx tele.Context) error {
	b.log.With("chat_id", teleCtx.Chat().ID, "msg_id", teleCtx.Message().ID).Debug("Sending help reply")
	return teleCtx.Send(b.cfg.MsgHelp)
}

func (b *bot) handle(teleCtx tele.Context) error {
	// ignore any non-Message updates
	if teleCtx.Message() == nil {
		return nil
	}

	// ignore group messages without bot mention
	// or without response to its previous message
	if !teleCtx.Message().Private() {
		if !b.isBotMention(teleCtx) && !b.isReplyToBot(teleCtx) {
			return nil
		}
	}

	return b.parseChatMessage(teleCtx)
}

func (b *bot) parseChatMessage(teleCtx tele.Context) error {
	log := b.log.With("chat_id", teleCtx.Chat().ID, "user_id", teleCtx.Sender().ID)
	log.Debug("Parsing new chat message", "message", teleCtx.Message().Text)
	return nil
}

func (b *bot) isBotMention(teleCtx tele.Context) bool {
	return strings.Contains(teleCtx.Message().Text, b.tg.Me.Username)
}

func (b *bot) isReplyToBot(teleCtx tele.Context) bool {
	reply := teleCtx.Message().ReplyTo
	if reply == nil || reply.Sender == nil {
		return false
	}
	return reply.Sender.ID == b.tg.Me.ID
}
