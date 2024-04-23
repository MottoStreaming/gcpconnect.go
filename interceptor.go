package gcpconnect

import (
	"log/slog"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/MottoStreaming/connectlog.go"
	"google.golang.org/protobuf/encoding/protojson"
)

// GetHandlerOptions returns the default options for a connect.Handler.
func GetHandlerOptions(logger *slog.Logger, opts ...Option) []connect.HandlerOption {
	o := options{
		// Default marshal options
		marshalOptions: protojson.MarshalOptions{
			// Use proto field names instead of JSON field names
			UseProtoNames: true,
			// Fill unpopulated fields with their default values
			EmitUnpopulated: true,
		},
	}
	for _, opt := range opts {
		opt(&o)
	}

	traceInt, err := otelconnect.NewInterceptor(
		// Disable metrics since they are producing a lot of data
		otelconnect.WithoutMetrics(),
		// Disable server peer attributes since they are producing high cardinality data
		otelconnect.WithoutServerPeerAttributes(),
		// Trust remote since most requests are coming through a proxy
		otelconnect.WithTrustRemote(),
	)
	if err != nil {
		logger.Error("failed to create tracing interceptor", "error", err)
	}

	return []connect.HandlerOption{
		connect.WithCodec(NewJSONCodec(o.marshalOptions)),
		connect.WithInterceptors(traceInt),
		connect.WithRecover(connectlog.NewLoggingRecoverHandler(logger)),
		// We log after recover so panic logs are not duplicated.
		// Internally, `connect.WithRecover` is adding interceptor.
		connect.WithInterceptors(connectlog.NewLoggingInterceptor(logger, o.logOptions...)),
	}
}
