package logic

import (
	logic_account "github.com/Fiagram/standalone/internal/logic/account"
	logic_chatbot "github.com/Fiagram/standalone/internal/logic/chatbot"
	logic_consumer "github.com/Fiagram/standalone/internal/logic/consumer"
	http_logic "github.com/Fiagram/standalone/internal/logic/http"
	token_logic "github.com/Fiagram/standalone/internal/logic/token"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"logic",
	fx.Provide(
		logic_account.NewHash,
		logic_account.NewAccount,

		token_logic.NewTokenLogic,

		http_logic.NewAuthLogic,
		http_logic.NewProfileLogic,
		http_logic.NewSubscriptionLogic,
		http_logic.NewStrategyLogic,

		logic_chatbot.NewCreatedWebhookChan,
		logic_chatbot.NewTorchSignalChan,
		logic_chatbot.NewWebhooksLogic,

		logic_consumer.NewTorchLogic,
	),
)
