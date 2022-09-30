package otlpexporter

import (
	"context"

	"github.com/TIQQE/aws-otel-lambda-layer/pkg/utility"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configcompression"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
)

const (
	// The value of "type" key in configuration.
	typeStr = "otlp"
)

// NewFactory creates a factory for OTLP exporter.
func NewFactory() component.ExporterFactory {
	return component.NewExporterFactory(
		typeStr,
		createDefaultConfig,
		component.WithTracesExporter(createTracesExporter, component.StabilityLevelStable),
		component.WithMetricsExporter(createMetricsExporter, component.StabilityLevelStable),
		component.WithLogsExporter(createLogsExporter, component.StabilityLevelBeta),
	)
}

func createDefaultConfig() config.Exporter {
	return &otlpexporter.Config{
		ExporterSettings: config.NewExporterSettings(config.NewComponentID(typeStr)),
		TimeoutSettings:  exporterhelper.NewDefaultTimeoutSettings(),
		RetrySettings:    exporterhelper.NewDefaultRetrySettings(),
		QueueSettings:    exporterhelper.NewDefaultQueueSettings(),
		GRPCClientSettings: configgrpc.GRPCClientSettings{
			Headers: map[string]string{},
			// Default to gzip compression
			Compression: configcompression.Gzip,
			// We almost read 0 bytes, so no need to tune ReadBufferSize.
			WriteBufferSize: 512 * 1024,
		},
	}
}

func createTracesExporter(ctx context.Context, set component.ExporterCreateSettings, cfg config.Exporter) (component.TracesExporter, error) {
	oce, err := newExporter(cfg, set)
	if err != nil {
		utility.LogError(err, "TracesExporterError", "Failed to create new traces exporter")
		return nil, err
	}

	oCfg := cfg.(*otlpexporter.Config)

	return exporterhelper.NewTracesExporter(ctx, set, cfg, oce.pushTraces,
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		exporterhelper.WithTimeout(oCfg.TimeoutSettings),
		exporterhelper.WithRetry(oCfg.RetrySettings),
		exporterhelper.WithQueue(oCfg.QueueSettings),
		exporterhelper.WithStart(oce.start),
		exporterhelper.WithShutdown(oce.shutdown))
}

func createMetricsExporter(ctx context.Context, set component.ExporterCreateSettings, cfg config.Exporter) (component.MetricsExporter, error) {
	oce, err := newExporter(cfg, set)
	if err != nil {
		utility.LogError(err, "MetricsExporterError", "Failed to create new metrics exporter")
		return nil, err
	}

	oCfg := cfg.(*otlpexporter.Config)

	return exporterhelper.NewMetricsExporter(ctx, set, cfg,
		oce.pushMetrics,
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		exporterhelper.WithTimeout(oCfg.TimeoutSettings),
		exporterhelper.WithRetry(oCfg.RetrySettings),
		exporterhelper.WithQueue(oCfg.QueueSettings),
		exporterhelper.WithStart(oce.start),
		exporterhelper.WithShutdown(oce.shutdown),
	)
}

func createLogsExporter(ctx context.Context, set component.ExporterCreateSettings, cfg config.Exporter) (component.LogsExporter, error) {
	oce, err := newExporter(cfg, set)
	if err != nil {
		utility.LogError(err, "LogsExporterError", "Failed to create new logs exporter")
		return nil, err
	}

	oCfg := cfg.(*otlpexporter.Config)

	return exporterhelper.NewLogsExporter(ctx, set, cfg,
		oce.pushLogs,
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		exporterhelper.WithTimeout(oCfg.TimeoutSettings),
		exporterhelper.WithRetry(oCfg.RetrySettings),
		exporterhelper.WithQueue(oCfg.QueueSettings),
		exporterhelper.WithStart(oce.start),
		exporterhelper.WithShutdown(oce.shutdown),
	)
}
