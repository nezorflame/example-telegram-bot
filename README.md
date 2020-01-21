# example-telegram-bot [![CircleCI](https://circleci.com/gh/nezorflame/example-telegram-bot/tree/master.svg?style=svg)](https://circleci.com/gh/nezorflame/example-telegram-bot/tree/master) [![Go Report Card](https://goreportcard.com/badge/github.com/nezorflame/example-telegram-bot)](https://goreportcard.com/report/github.com/nezorflame/example-telegram-bot) [![GolangCI](https://golangci.com/badges/github.com/nezorflame/example-telegram-bot.svg)](https://golangci.com/r/github.com/nezorflame/example-telegram-bot) [![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fnezorflame%2Fexample-telegram-bot.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fnezorflame%2Fexample-telegram-bot?ref=badge_shield)

Example bot template for Telegram.

## Description

With this type of setup all you need to do is:

- `go get` the bot (or `git clone` it)
- add required code
- change the config file to your needs
- modify `.service` file for systemd to manage your bot
- deploy your bot to the server of choice, using modified config and service files

## Dependencies

This bot uses:

- [tgbotapi](github.com/go-telegram-bot-api/telegram-bot-api) package to work with Telegram API
- [bbolt](go.etcd.io/bbolt) for local database
- [viper](github.com/spf13/viper) for configuration and [pflag](github.com/spf13/pflag) for command flags
- [logrus](github.com/sirupsen/logrus) for logging

## Structure

This project adheres to the golang-standards [Standard Go Project Layout](https://github.com/golang-standards/project-layout) structure:

- `internal/pkg` holds the private libraries:
  - `config` for configuration
  - `db` for database
  - `file` for file and network helpers
- `pkg` holds the public libraries (mainly `telegram` package with bot implementation)

## Customization

To add another custom command handler, you can:

- add a command to `config.toml` (and also a corresponding message, if required)
- edit `internal/pkg`

## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fnezorflame%2Fexample-telegram-bot.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fnezorflame%2Fexample-telegram-bot?ref=badge_large)
