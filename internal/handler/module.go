package handler

import (
	"context"
	"fmt"

	handler "github.com/Fiagram/standalone/internal/handler/http"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"handler",
	fx.Provide(
		handler.NewHttpServer,
	),
	fx.Invoke(
		func(lc fx.Lifecycle, s handler.HttpServer) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go s.Start(ctx)
					return nil
				},
				OnStop: func(ctx context.Context) error {
					fmt.Println("Stopping server")
					return nil
				},
			})
		},
	),
)
