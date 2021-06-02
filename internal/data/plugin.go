package data

import (
	"fmt"

	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel"

	oteltrace "go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

const (
	defaultTracerName  = "go.opentelemetry.io/contrib/instrumentation/github.com/go-gorm/gorm/otelgorm"
	defaultServiceName = "gorm"

	callBackBeforeName = "otel:before"
	callBackAfterName  = "otel:after"
)

type gormHookFunc func(tx *gorm.DB)

type OtelPlugin struct {
	serviceName    string
	tracerProvider oteltrace.TracerProvider
	tracer         oteltrace.Tracer
}

func (op *OtelPlugin) Name() string {
	return "OpenTelemetryPlugin"
}

// NewPlugin initialize a new gorm.DB plugin that traces queries
// You may pass optional Options to the function
func NewPlugin() *OtelPlugin {
	provider := otel.GetTracerProvider()
	return &OtelPlugin{
		serviceName:    "",
		tracerProvider: nil,
		tracer: provider.Tracer(
			defaultTracerName,
			oteltrace.WithInstrumentationVersion(contrib.SemVersion()),
		),
	}
}

type registerCallback interface {
	Register(name string, fn func(*gorm.DB)) error
}

func beforeName(name string) string {
	return callBackBeforeName + "_" + name
}

func afterName(name string) string {
	return callBackAfterName + "_" + name
}

func (op *OtelPlugin) Initialize(db *gorm.DB) error {
	registerHooks := []struct {
		callback registerCallback
		hook     gormHookFunc
		name     string
	}{
		// before hooks
		{db.Callback().Create().Before("gorm:before_create"), op.before, beforeName("create")},
		{db.Callback().Query().Before("gorm:query"), op.before, beforeName("query")},
		{db.Callback().Delete().Before("gorm:before_delete"), op.before, beforeName("delete")},
		{db.Callback().Update().Before("gorm:before_update"), op.before, beforeName("update")},
		{db.Callback().Row().Before("gorm:row"), op.before, beforeName("row")},
		{db.Callback().Raw().Before("gorm:raw"), op.before, beforeName("raw")},

		// after hooks
		{db.Callback().Create().After("gorm:after_create"), op.after("INSERT"), afterName("create")},
		{db.Callback().Query().After("gorm:after_query"), op.after("SELECT"), afterName("select")},
		{db.Callback().Delete().After("gorm:after_delete"), op.after("DELETE"), afterName("delete")},
		{db.Callback().Update().After("gorm:after_update"), op.after("UPDATE"), afterName("update")},
		{db.Callback().Row().After("gorm:row"), op.after(""), afterName("row")},
		{db.Callback().Raw().After("gorm:raw"), op.after(""), afterName("raw")},
	}

	for _, h := range registerHooks {
		if err := h.callback.Register(h.name, h.hook); err != nil {
			return fmt.Errorf("register %s hook: %w", h.name, err)
		}
	}

	return nil
}
