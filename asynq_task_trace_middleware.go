package connect

import (
	"context"
	"fmt"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// AsynqTaskTracerMiddleware tracer for asynq task, place this middleware on mux
func AsynqTaskTracerMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		tracer := otel.GetTracerProvider().Tracer(
			instrumentationName,
		)
		ctx, span := tracer.Start(ctx, fmt.Sprintf("asynq-tracer-task-%s", t.Type()))
		span.SetAttributes(attribute.String("task.type", t.Type()))
		span.SetAttributes(attribute.String("task.payload", string(t.Payload())))
		defer span.End()

		return h.ProcessTask(ctx, t)
	})
}
