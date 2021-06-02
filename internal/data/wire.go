// +build wireinject

// The build tag makes sure the stub is not built in the final build.
//go:generate  wire
package data

import (
	"github.com/go-kratos/kratos-layout/internal/conf"
	"io"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

func NewData(conf *conf.Data, l log.Logger) (*Data, func(), error) {
	panic(wire.Build(log.NewHelper, NewMysql, NewRedis, NewMongoDB, wire.Struct(new(Data), "*")))
}

func newTestRepo(*conf.Data) (*greeterRepo, func(), error) {
	panic(wire.Build(wire.InterfaceValue(new(io.Writer), os.Stdout), log.NewStdLogger, NewData, newGreeterRepo))
}
