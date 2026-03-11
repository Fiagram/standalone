package app

import (
	"github.com/Fiagram/standalone/internal/configs"
	"github.com/Fiagram/standalone/internal/dao"
	"github.com/Fiagram/standalone/internal/handler"
	"github.com/Fiagram/standalone/internal/logger"
	"github.com/Fiagram/standalone/internal/logic"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"app",
	configs.Module,
	logger.Module,

	dao.Module,

	logic.Module,
	handler.Module,
)
