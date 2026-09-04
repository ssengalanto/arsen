package server

import (
	"context"
	"log/slog"
	"net"
	"net/http"

	"go.uber.org/fx"
)

// Module provides the HTTP server to the fx dependency graph and manages its
// lifecycle. It uses net.Listen so the port is bound before OnStart returns,
// which lets fx know the server is truly ready.
var Module = fx.Module("server",
	fx.Provide(New),
	fx.Invoke(func(lc fx.Lifecycle, srv *Server) {
		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				ln, err := net.Listen("tcp", srv.HTTP.Addr)
				if err != nil {
					return err
				}
				slog.Info("server started", "address", srv.HTTP.Addr)
				go func() {
					if err := srv.HTTP.Serve(ln); err != nil && err != http.ErrServerClosed {
						slog.Error("server exited unexpectedly", "error", err)
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				slog.Info("server shutting down")
				return srv.HTTP.Shutdown(ctx)
			},
		})
	}),
)
