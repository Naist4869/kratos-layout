package main

import (
	"flag"
	"github.com/go-kratos/kratos-layout/internal/data"
	"github.com/go-kratos/kratos-layout/pkg/config"
	"github.com/go-kratos/kratos-layout/pkg/decodehook"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"os"

	"github.com/go-kratos/kratos-layout/internal/conf"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	_ "github.com/go-sql-driver/mysql"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string

	BuildTime string
	// flagconf is the config flag.
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "./conf", "config path, eg: -conf application.yaml")
}

func newApp(logger log.Logger, hs *http.Server, gs *grpc.Server, data *data.Data) *kratos.App {
	return kratos.New(
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			hs,
			gs,
			data,
		),
	)
}

func main() {
	flag.Parse()
	env, exist := os.LookupEnv("DEPLOY_ENV")
	logger := log.With(log.NewStdLogger(os.Stdout),
		"service.name", Name,
		"service.version", Version,
		"service.buildTime", BuildTime,
		"service.deploy", env,
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
	)
	if err := config.Init("", Name); err != nil {
		panic(err)
	}

	s := new(conf.Server)
	d := new(conf.Data)
	t := new(conf.OTEL)
	if err := viper.UnmarshalKey("server", s, viper.DecodeHook(decodehook.StringToTimeDurationHookFunc())); err != nil {
		panic(err)
	}
	if err := viper.UnmarshalKey("data", d, viper.DecodeHook(decodehook.StringToTimeDurationHookFunc())); err != nil {
		panic(err)
	}
	if err := viper.UnmarshalKey("otel", t, viper.DecodeHook(decodehook.StringToTimeDurationHookFunc())); err != nil {
		panic(err)
	}

	var tp *sdktrace.TracerProvider

	if exporter, err := jaeger.NewRawExporter(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(t.CollectorEndpoint))); err != nil {
		panic(err)
	} else if exist && env != "production" {
		tp = sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter), sdktrace.WithSampler(sdktrace.AlwaysSample()))
	} else {
		tp = sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter), sdktrace.WithSampler(sdktrace.NeverSample()))
	}

	app, cleanup, err := initApp(s, d, tp, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
