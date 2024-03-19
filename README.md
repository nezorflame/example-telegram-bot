# example-telegram-bot [![Workflow status](https://github.com/nezorflame/example-telegram-bot/actions/workflows/go.yml/badge.svg)](https://github.com/nezorflame/example-telegram-bot/actions/workflows/go.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/nezorflame/example-telegram-bot)](https://goreportcard.com/report/github.com/nezorflame/example-telegram-bot) [![GolangCI](https://golangci.com/badges/github.com/nezorflame/example-telegram-bot.svg)](https://golangci.com/r/github.com/nezorflame/example-telegram-bot) [![FOSSA license check](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fnezorflame%2Fexample-telegram-bot.svg?type=shield&issueType=license)](https://app.fossa.com/projects/git%2Bgithub.com%2Fnezorflame%2Fexample-telegram-bot?ref=badge_shield&issueType=license) [![FOSSA security check](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fnezorflame%2Fexample-telegram-bot.svg?type=shield&issueType=security)](https://app.fossa.com/projects/git%2Bgithub.com%2Fnezorflame%2Fexample-telegram-bot?ref=badge_shield&issueType=security)

Example bot template for Telegram.

## Description

With this type of setup all you need to do is:

- create a project from the template and `git clone` it
- replace the module and bot name to your own
- run `make init` to set up the project and its dependencies
- add required code
- set up the environment or your `.env` file for your needs
- modify `.service` file for systemd to manage your bot
- deploy your bot to the server of choice!

## Dependencies

This bot uses:

- [telebot](https://pkg.go.dev/gopkg.in/telebot.v3) package to work with Telegram API
- [bolt](https://pkg.go.dev/go.etcd.io/bbolt) for the database
- [envconfig](https://pkg.go.dev/github.com/kelseyhightower/envconfig) + [godotenv](https://pkg.go.dev/github.com/joho/godotenv) for the configuration
- [slog](https://pkg.go.dev/log/slog) for the logging

## Structure

This project mostly adheres to the [Project Layout](https://github.com/golang-standards/project-layout) structure, excluding `pkg` folders.

`internal` package holds the private libraries:

- `config` for configuration
- `bolt` for database (using BoltDB)
- `file` for file and network helpers
- `telegram` package with bot implementation

## Customization

To add another custom command handler, you can:

- add a command to `.env` file (and also a corresponding message, if required)
- edit `internal` packages

## License

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fnezorflame%2Fexample-telegram-bot.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fnezorflame%2Fexample-telegram-bot?ref=badge_large)
