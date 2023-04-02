package bot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nezorflame/example-telegram-bot/internal/bolt"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

type bot struct {
	api *tgbotapi.BotAPI
	db  *bolt.DB

	cfg *viper.Viper
	log *slog.Logger
}

// New creates new instance of Bot
func New(cfg *viper.Viper, log *slog.Logger, logLevel slog.Level, db *bolt.DB) (*bot, error) {
	if cfg == nil {
		return nil, errors.New("empty config")
	}
	if log == nil {
		return nil, errors.New("empty logger")
	}
	if db == nil {
		return nil, errors.New("empty DB")
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
	return &bot{api: api, db: db, cfg: cfg, log: log}, nil
}

// Start starts to listen the bot updates channel
func (b *bot) Start(ctx context.Context) {
	update := tgbotapi.NewUpdate(0)
	update.Timeout = b.cfg.GetInt("telegram.timeout")
	updates := b.api.GetUpdatesChan(update)
	go b.listen(ctx, updates)
}

// Stop stops the bot
func (b *bot) Stop() {
	b.api.StopReceivingUpdates()
}

func (b *bot) listen(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	for {
		select {
		case <-ctx.Done():
			b.log.Warn("Context closed - stopping listening to the updates", "error", ctx.Err())
			return
		case u := <-updates:
			// ignore any non-Message updates
			if u.Message == nil {
				continue
			}

			// ignore group messages without bot mention
			// or without response to its previous message
			if !u.FromChat().IsPrivate() {
				if !b.isBotMention(u.Message) && !b.isReplyToBot(u.Message) {
					continue
				}
			}

			switch {
			case strings.EqualFold(u.Message.Command(), b.cfg.GetString("commands.start")):
				b.log.With("user_id", u.Message.From.ID).Debug("Got /start command")
				fallthrough
			case strings.EqualFold(u.Message.Command(), b.cfg.GetString("commands.help")):
				b.log.With("user_id", u.Message.From.ID).Debug("Got /help command")
				go b.help(u.Message)
			default:
				b.log.With("user_id", u.Message.From.ID).Debug("Got a message")
				go b.parseChatMessage(ctx, u.Message)
			}
		}
	}
}

func (b *bot) help(msg *tgbotapi.Message) {
	b.reply(msg.Chat.ID, msg.MessageID, b.cfg.GetString("messages.help"))
}

func (b *bot) reply(chatID int64, msgID int, text string) {
	b.log.With("chat_id", chatID).With("msg_id", msgID).Debug("Sending reply")
	msg := tgbotapi.NewMessage(chatID, fmt.Sprint("", text))
	if msgID != 0 {
		msg.ReplyToMessageID = msgID
	}
	msg.ParseMode = tgbotapi.ModeMarkdown

	if _, err := b.api.Send(msg); err != nil {
		b.log.Error("Unable to send the message", "error", err)
		return
	}
}

func (b *bot) parseChatMessage(ctx context.Context, msg *tgbotapi.Message) {
	log := b.log.With("chat_id", strconv.FormatInt(msg.Chat.ID, 10), "user_id", msg.From.ID)
	log.DebugCtx(ctx, "Parsing new chat message", "message", msg.Text)
}

func (b *bot) isBotMention(msg *tgbotapi.Message) bool {
	return strings.Contains(msg.Text, b.api.Self.UserName)
}

func (b *bot) isReplyToBot(msg *tgbotapi.Message) bool {
	if msg.ReplyToMessage == nil || msg.ReplyToMessage.From == nil {
		return false
	}
	return msg.ReplyToMessage.From.ID == b.api.Self.ID
}
