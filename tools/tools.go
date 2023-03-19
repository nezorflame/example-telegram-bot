//go:build tools

//go:generate go build -o ../bin/mockgen github.com/golang/mock/mockgen
//go:generate bash -c "go build -ldflags \"-X 'main.version=$(go list -m -f '{{.Version}}' github.com/golangci/golangci-lint)' -X 'main.commit=test' -X 'main.date=test'\" -o ../bin/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint"
//go:generate go build -o ../bin/gosimports github.com/rinchsan/gosimports/cmd/gosimports
//go:generate go build -o ../bin/gofumpt mvdan.cc/gofumpt

// Package tools contains go:generate commands for all project tools with versions stored in local go.mod file
// See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
package tools

import (
	_ "github.com/golang/mock/mockgen"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/rinchsan/gosimports/cmd/gosimports"
	_ "mvdan.cc/gofumpt"
)
