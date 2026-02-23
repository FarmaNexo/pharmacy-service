// internal/presentation/http/middlewares/correlation_id.go
package middlewares

import (
	"context"
	"net/http"

	"github.com/farmanexo/pharmacy-service/pkg/mediator"
	"github.com/google/uuid"
)

func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		w.Header().Set("X-Correlation-ID", correlationID)

		ctx := mediator.WithValue(r.Context(), mediator.CorrelationKey, correlationID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetCorrelationID(ctx context.Context) string {
	if val := ctx.Value(mediator.CorrelationKey); val != nil {
		if corrID, ok := val.(string); ok {
			return corrID
		}
	}
	return ""
}
