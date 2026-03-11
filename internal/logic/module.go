package logic

import (
	logic_account "github.com/Fiagram/standalone/internal/logic/account"
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
		http_logic.NewUsersLogic,
	),
)
