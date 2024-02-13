package http

import (
	"github.com/felixge/httpsnoop"
	"github.com/go-chi/chi/v5"
	"gitlab.tixlabs.io/apps/common/http/constants"
	"gitlab.tixlabs.io/apps/common/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"sync"
)

const (
	// instrumentationName is the name of this instrumentation package.
	instrumentationName = "gitlab.tixlabs.io/apps/common/tracing/http"
)

func Middleware() func(next http.Handler) http.Handler {
	cfg := tracing.GetConfig()

	if !cfg.Initialized {
		panic("Tracer is not initialized")
	}

	tracer := cfg.TracerProvider.Tracer(
		instrumentationName,
	)
	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}
	return func(handler http.Handler) http.Handler {
		return traceware{
			serverName:  cfg.ServiceName,
			tracer:      tracer,
			propagators: cfg.Propagators,
			handler:     handler,
		}
	}
}

type traceware struct {
	serverName  string
	tracer      trace.Tracer
	propagators propagation.TextMapPropagator
	handler     http.Handler
}

type recordingResponseWriter struct {
	writer  http.ResponseWriter
	written bool
	status  int
}

var rrwPool = &sync.Pool{
	New: func() interface{} {
		return &recordingResponseWriter{}
	},
}

func getRRW(writer http.ResponseWriter) *recordingResponseWriter {
	rrw := rrwPool.Get().(*recordingResponseWriter)
	rrw.written = false
	rrw.status = 0
	rrw.writer = httpsnoop.Wrap(writer, httpsnoop.Hooks{
		Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
			return func(b []byte) (int, error) {
				if !rrw.written {
					rrw.written = true
					rrw.status = http.StatusOK
				}
				return next(b)
			}
		},
		WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return func(statusCode int) {
				if !rrw.written {
					rrw.written = true
					rrw.status = statusCode
				}
				next(statusCode)
			}
		},
	})
	return rrw
}

func putRRW(rrw *recordingResponseWriter) {
	rrw.writer = nil
	rrwPool.Put(rrw)
}

// ServeHTTP implements the http.Handler interface. It does the actual
// tracing of the request.
func (tw traceware) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path == constants.PATH_HEALTHZ || r.URL.Path == constants.PATH_READYZ {
		tw.handler.ServeHTTP(w, r)
		return
	}

	// extract tracing header using propagator
	ctx := tw.propagators.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

	spanName := ""
	routePattern := ""

	ctx, span := tracing.StartSpanFromContext(
		trace.ContextWithRemoteSpanContext(ctx, trace.SpanContextFromContext(ctx)),
		spanName,
		trace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", r)...),
		trace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(r)...),
		trace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(tw.serverName, routePattern, r)...),
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	// get recording response writer
	rrw := getRRW(w)
	defer putRRW(rrw)

	// execute next http handler
	r = r.WithContext(ctx)
	tw.handler.ServeHTTP(rrw.writer, r)

	// set span name & http route attribute if necessary
	routePattern = chi.RouteContext(r.Context()).RoutePattern()
	span.SetAttributes(semconv.HTTPRouteKey.String(routePattern))

	span.SetName(routePattern)

	// set status code attribute
	span.SetAttributes(semconv.HTTPStatusCodeKey.Int(rrw.status))

	// set span status
	spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCode(rrw.status)
	span.SetStatus(spanStatus, spanMessage)
}
