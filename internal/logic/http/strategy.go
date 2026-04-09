package logic_http

import (
	"net/http"

	dao_strategy "github.com/Fiagram/standalone/internal/dao/strategy"
	pb "github.com/Fiagram/standalone/internal/generated/grpc/strategy"
	oapi "github.com/Fiagram/standalone/internal/generated/openapi"
	"github.com/Fiagram/standalone/internal/logger"
	"github.com/Fiagram/standalone/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type StrategyLogic interface {
	GetStrategyAlerts(c *gin.Context, params oapi.GetStrategyAlertsParams)
	CreateStrategyAlert(c *gin.Context)
	GetStrategyAlert(c *gin.Context, alertId oapi.AlertId)
	UpdateStrategyAlert(c *gin.Context, alertId oapi.AlertId)
	DeleteStrategyAlert(c *gin.Context, alertId oapi.AlertId)
}

var _ StrategyLogic = (oapi.ServerInterface)(nil)

type strategyLogic struct {
	strategyClient dao_strategy.Client
	logger         *zap.Logger
}

func NewStrategyLogic(
	strategyClient dao_strategy.Client,
	logger *zap.Logger,
) StrategyLogic {
	return &strategyLogic{
		strategyClient: strategyClient,
		logger:         logger,
	}
}

// ──────────────────────────────────────────────
// Enum mapping helpers: oapi ↔ gRPC
// ──────────────────────────────────────────────

var timeframeToProto = map[oapi.Timeframe]pb.Alert_Timeframe{
	oapi.D1: pb.Alert_TIMEFRAME_D1,
	oapi.W1: pb.Alert_TIMEFRAME_W1,
	oapi.M1: pb.Alert_TIMEFRAME_M1,
}

var timeframeFromProto = map[pb.Alert_Timeframe]oapi.Timeframe{
	pb.Alert_TIMEFRAME_D1: oapi.D1,
	pb.Alert_TIMEFRAME_W1: oapi.W1,
	pb.Alert_TIMEFRAME_M1: oapi.M1,
}

var indicatorToProto = map[oapi.Indicator]pb.Alert_Indicator{
	oapi.Close:          pb.Alert_INDICATOR_CLOSE,
	oapi.BollingerBands: pb.Alert_INDICATOR_BOLLINGER_BANDS,
	oapi.Rsi:            pb.Alert_INDICATOR_RSI,
	oapi.Ma10:           pb.Alert_INDICATOR_MA10,
	oapi.Ma50:           pb.Alert_INDICATOR_MA50,
	oapi.Ma100:          pb.Alert_INDICATOR_MA100,
	oapi.Ma200:          pb.Alert_INDICATOR_MA200,
}

var indicatorFromProto = map[pb.Alert_Indicator]oapi.Indicator{
	pb.Alert_INDICATOR_CLOSE:           oapi.Close,
	pb.Alert_INDICATOR_BOLLINGER_BANDS: oapi.BollingerBands,
	pb.Alert_INDICATOR_RSI:             oapi.Rsi,
	pb.Alert_INDICATOR_MA10:            oapi.Ma10,
	pb.Alert_INDICATOR_MA50:            oapi.Ma50,
	pb.Alert_INDICATOR_MA100:           oapi.Ma100,
	pb.Alert_INDICATOR_MA200:           oapi.Ma200,
}

var operatorToProto = map[oapi.Operator]pb.Alert_Operator{
	oapi.GreaterThan:  pb.Alert_OPERATOR_GREATER_THAN,
	oapi.LessThan:     pb.Alert_OPERATOR_LESS_THAN,
	oapi.CrossingUp:   pb.Alert_OPERATOR_CROSSING_UP,
	oapi.CrossingDown: pb.Alert_OPERATOR_CROSSING_DOWN,
	oapi.Crossing:     pb.Alert_OPERATOR_CROSSING,
}

var operatorFromProto = map[pb.Alert_Operator]oapi.Operator{
	pb.Alert_OPERATOR_GREATER_THAN:  oapi.GreaterThan,
	pb.Alert_OPERATOR_LESS_THAN:     oapi.LessThan,
	pb.Alert_OPERATOR_CROSSING_UP:   oapi.CrossingUp,
	pb.Alert_OPERATOR_CROSSING_DOWN: oapi.CrossingDown,
	pb.Alert_OPERATOR_CROSSING:      oapi.Crossing,
}

var triggerToProto = map[oapi.Trigger]pb.Alert_Trigger{
	oapi.Once:  pb.Alert_TRIGGER_ONCE,
	oapi.Every: pb.Alert_TRIGGER_EVERY,
}

var triggerFromProto = map[pb.Alert_Trigger]oapi.Trigger{
	pb.Alert_TRIGGER_ONCE:  oapi.Once,
	pb.Alert_TRIGGER_EVERY: oapi.Every,
}

// alertFromProto converts a gRPC Alert to an OpenAPI Alert.
func alertFromProto(a *pb.Alert) oapi.Alert {
	out := oapi.Alert{
		Id:        utils.Ptr(a.Id),
		Timeframe: timeframeFromProto[a.Timeframe],
		Symbol:    a.Symbol,
		Indicator: indicatorFromProto[a.Indicator],
		Operator:  operatorFromProto[a.Operator],
		Trigger:   triggerFromProto[a.Trigger],
		Exp:       a.Exp,
		Message:   a.Message,
	}
	if a.CreatedAt != nil {
		t := a.CreatedAt.AsTime()
		out.CreatedAt = &t
	}
	if a.UpdatedAt != nil {
		t := a.UpdatedAt.AsTime()
		out.UpdatedAt = &t
	}
	return out
}

// ──────────────────────────────────────────────
// Handlers
// ──────────────────────────────────────────────

func (s *strategyLogic) GetStrategyAlerts(c *gin.Context, params oapi.GetStrategyAlertsParams) {
	logger := logger.LoggerWithContext(c, s.logger)

	accountId, ok := getAccountIdFromContext(c, logger)
	if !ok {
		return
	}

	limit := uint32(20)
	offset := uint32(0)
	if params.Limit != nil {
		limit = uint32(*params.Limit)
	}
	if params.Offset != nil {
		offset = uint32(*params.Offset)
	}

	resp, err := s.strategyClient.GetAlerts(c, &pb.GetAlertsRequest{
		OfAccountId: accountId,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		errMsg := "failed to get alerts from strategy service"
		logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	out := make([]oapi.Alert, 0, len(resp.Alerts))
	for _, a := range resp.Alerts {
		out = append(out, alertFromProto(a))
	}

	c.JSON(http.StatusOK, out)
}

func (s *strategyLogic) CreateStrategyAlert(c *gin.Context) {
	logger := logger.LoggerWithContext(c, s.logger)

	accountId, ok := getAccountIdFromContext(c, logger)
	if !ok {
		return
	}

	var req oapi.CreateStrategyAlertJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "invalid request body",
		})
		return
	}

	resp, err := s.strategyClient.CreateAlert(c, &pb.CreateAlertRequest{
		OfAccountId: accountId,
		Timeframe:   timeframeToProto[req.Timeframe],
		Symbol:      req.Symbol,
		Indicator:   indicatorToProto[req.Indicator],
		Operator:    operatorToProto[req.Operator],
		Trigger:     triggerToProto[req.Trigger],
		Exp:         req.Exp,
		Message:     req.Message,
	})
	if err != nil {
		errMsg := "failed to create alert in strategy service"
		logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	c.JSON(http.StatusCreated, alertFromProto(resp.Alert))
}

func (s *strategyLogic) GetStrategyAlert(c *gin.Context, alertId oapi.AlertId) {
	logger := logger.LoggerWithContext(c, s.logger)

	accountId, ok := getAccountIdFromContext(c, logger)
	if !ok {
		return
	}

	resp, err := s.strategyClient.GetAlert(c, &pb.GetAlertRequest{
		OfAccountId: accountId,
		AlertId:     alertId,
	})
	if err != nil {
		errMsg := "alert not found"
		logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusNotFound, oapi.NotFound{
			Code:    "NotFound",
			Message: errMsg,
		})
		return
	}

	c.JSON(http.StatusOK, alertFromProto(resp.Alert))
}

func (s *strategyLogic) UpdateStrategyAlert(c *gin.Context, alertId oapi.AlertId) {
	logger := logger.LoggerWithContext(c, s.logger)

	accountId, ok := getAccountIdFromContext(c, logger)
	if !ok {
		return
	}

	var req oapi.UpdateStrategyAlertJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "invalid request body",
		})
		return
	}

	resp, err := s.strategyClient.UpdateAlert(c, &pb.UpdateAlertRequest{
		OfAccountId: accountId,
		AlertId:     alertId,
		Timeframe:   timeframeToProto[req.Timeframe],
		Symbol:      req.Symbol,
		Indicator:   indicatorToProto[req.Indicator],
		Operator:    operatorToProto[req.Operator],
		Trigger:     triggerToProto[req.Trigger],
		Exp:         req.Exp,
		Message:     req.Message,
	})
	if err != nil {
		errMsg := "failed to update alert in strategy service"
		logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	c.JSON(http.StatusOK, alertFromProto(resp.Alert))
}

func (s *strategyLogic) DeleteStrategyAlert(c *gin.Context, alertId oapi.AlertId) {
	logger := logger.LoggerWithContext(c, s.logger)

	accountId, ok := getAccountIdFromContext(c, logger)
	if !ok {
		return
	}

	_, err := s.strategyClient.DeleteAlert(c, &pb.DeleteAlertRequest{
		OfAccountId: accountId,
		AlertId:     alertId,
	})
	if err != nil {
		errMsg := "failed to delete alert in strategy service"
		logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: errMsg,
		})
		return
	}

	c.Status(http.StatusNoContent)
}
