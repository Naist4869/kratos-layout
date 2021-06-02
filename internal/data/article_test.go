package data

import (
	"context"
	"github.com/go-kratos/kratos-layout/internal/conf"
	"github.com/go-kratos/kratos-layout/pkg/config"
	"github.com/go-kratos/kratos-layout/pkg/decodehook"
	"os"
	"testing"

	"github.com/spf13/viper"
)

var testRepo *greeterRepo
var ctx = context.Background()

func TestMain(m *testing.M) {
	testDirPath, err := config.GuessTestDirPath()
	if err != nil {
		panic(err)
	}
	if err = config.Init(testDirPath+"application.yaml", ""); err != nil {
		panic(err)
	}
	data := new(conf.Data)
	if err := viper.UnmarshalKey("data", data, viper.DecodeHook(decodehook.StringToTimeDurationHookFunc())); err != nil {
		panic(err)
	}
	var clear func()
	testRepo, clear, err = newTestRepo(data)
	if err != nil {
		panic(err)
	}
	ret := m.Run()
	clear()
	os.Exit(ret)
}
