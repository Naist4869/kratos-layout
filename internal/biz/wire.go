// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package biz

//go:generate wire

import (
	"io"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// newBiz init kratos application.
func newBiz(repo GreeterRepo) (*GreeterUsecase, error) {
	panic(wire.Build(wire.InterfaceValue(new(io.Writer), os.Stdout), log.NewStdLogger, ProviderSet))
}
