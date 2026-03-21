package handler

import (
	"context"
	"fmt"

	webhook_handler "github.com/Fiagram/standalone/internal/handler/chatbot"
	http_handler "github.com/Fiagram/standalone/internal/handler/http"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"handler",
	fx.Provide(
		http_handler.NewHttpServer,
		webhook_handler.NewCreatedWebhookChan,
		webhook_handler.NewWebhookServer,
	),
	fx.Invoke(
		func(lc fx.Lifecycle,
			hs http_handler.HttpServer,
			ws webhook_handler.WebhookServer,
		) {
			var cancel context.CancelFunc

			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					var ctx context.Context
					ctx, cancel = context.WithCancel(context.Background())
					go hs.Start(ctx)
					go ws.Start(ctx)
					return nil
				},
				OnStop: func(_ context.Context) error {
					fmt.Println("Stopping server")
					cancel()
					return nil
				},
			})
		},
	),
)
