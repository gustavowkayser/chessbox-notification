package middleware

import (
	"chessbox-notification/internal/utils"
	"context"
	"net/http"
	"strings"
)

type Device struct {
	ID string
	CredentialHash string
	PushToken string
	Platform string
	AppVersion string
	Active bool
}

type Service interface {
	Authenticate(context.Context, string) (*Device, error)
}

func DeviceFromContext(ctx context.Context) *Device {
	device, ok := ctx.Value("device").(*Device)

	if !ok {
		return nil
	}

	return device
}

func DeviceAuth(authService Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			wr := utils.NewWriteReader[any, any](r, w)
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

			if token == "" {
				wr.WriteError(
					http.StatusUnauthorized,
					"Unauthorized",
					"Invalid token",
				)
				return
			}

			device, err := authService.Authenticate(r.Context(), token)
			if err != nil {
				wr.WriteError(
					http.StatusUnauthorized,
					"Unauthorized",
					"Invalid token",
				)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				"device",
				device,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
