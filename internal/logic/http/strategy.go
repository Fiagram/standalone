package logic_http

import (
	"encoding/json"
	"fmt"
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

var timeframeToProto = map[oapi.Timeframe]pb.Timeframe{
	oapi.D1: pb.Timeframe_TIMEFRAME_D1,
	oapi.W1: pb.Timeframe_TIMEFRAME_W1,
	oapi.M1: pb.Timeframe_TIMEFRAME_M1,
}

var timeframeFromProto = map[pb.Timeframe]oapi.Timeframe{
	pb.Timeframe_TIMEFRAME_D1: oapi.D1,
	pb.Timeframe_TIMEFRAME_W1: oapi.W1,
	pb.Timeframe_TIMEFRAME_M1: oapi.M1,
}

var operatorToProto = map[oapi.Operator]pb.Operator{
	oapi.GreaterThan:  pb.Operator_OPERATOR_GREATER_THAN,
	oapi.LessThan:     pb.Operator_OPERATOR_LESS_THAN,
	oapi.CrossingUp:   pb.Operator_OPERATOR_CROSSING_UP,
	oapi.CrossingDown: pb.Operator_OPERATOR_CROSSING_DOWN,
	oapi.Crossing:     pb.Operator_OPERATOR_CROSSING,
}

var operatorFromProto = map[pb.Operator]oapi.Operator{
	pb.Operator_OPERATOR_GREATER_THAN:  oapi.GreaterThan,
	pb.Operator_OPERATOR_LESS_THAN:     oapi.LessThan,
	pb.Operator_OPERATOR_CROSSING_UP:   oapi.CrossingUp,
	pb.Operator_OPERATOR_CROSSING_DOWN: oapi.CrossingDown,
	pb.Operator_OPERATOR_CROSSING:      oapi.Crossing,
}

var triggerToProto = map[oapi.Trigger]pb.Trigger{
	oapi.Once:  pb.Trigger_TRIGGER_ONCE,
	oapi.Every: pb.Trigger_TRIGGER_EVERY,
}

var triggerFromProto = map[pb.Trigger]oapi.Trigger{
	pb.Trigger_TRIGGER_ONCE:  oapi.Once,
	pb.Trigger_TRIGGER_EVERY: oapi.Every,
}

// ──────────────────────────────────────────────
// Operand conversion helpers: oapi ↔ gRPC
// ──────────────────────────────────────────────

// priceFromProto and friends map proto enum values to their OpenAPI string representations.
var priceFromProto = map[pb.Price]string{
	pb.Price_PRICE_OPEN:  "open",
	pb.Price_PRICE_HIGH:  "high",
	pb.Price_PRICE_LOW:   "low",
	pb.Price_PRICE_CLOSE: "close",
}

var bbFromProto = map[pb.BollingerBand]string{
	pb.BollingerBand_BOLLINGER_BAND_MIDDLE: "middle",
	pb.BollingerBand_BOLLINGER_BAND_LOWER:  "lower",
	pb.BollingerBand_BOLLINGER_BAND_UPPER:  "upper",
}

var smaFromProto = map[pb.SimpleMovingAverage]string{
	pb.SimpleMovingAverage_SIMPLE_MOVING_AVERAGE_SMA10:  "sma10",
	pb.SimpleMovingAverage_SIMPLE_MOVING_AVERAGE_SMA50:  "sma50",
	pb.SimpleMovingAverage_SIMPLE_MOVING_AVERAGE_SMA100: "sma100",
	pb.SimpleMovingAverage_SIMPLE_MOVING_AVERAGE_SMA200: "sma200",
}

// protoOperandToJSON serialises a proto Operand to its OpenAPI JSON wire form
// (a JSON string for indicator-based operands, a JSON number for const values).
func protoOperandToJSON(o *pb.Operand) (json.RawMessage, error) {
	if o == nil {
		return nil, fmt.Errorf("nil operand")
	}
	var v any
	switch val := o.Value.(type) {
	case *pb.Operand_Price:
		s, ok := priceFromProto[val.Price]
		if !ok {
			return nil, fmt.Errorf("unknown Price enum value: %v", val.Price)
		}
		v = s
	case *pb.Operand_BollingerBand:
		s, ok := bbFromProto[val.BollingerBand]
		if !ok {
			return nil, fmt.Errorf("unknown BollingerBand enum value: %v", val.BollingerBand)
		}
		v = s
	case *pb.Operand_SimpleMovingAverage:
		s, ok := smaFromProto[val.SimpleMovingAverage]
		if !ok {
			return nil, fmt.Errorf("unknown SimpleMovingAverage enum value: %v", val.SimpleMovingAverage)
		}
		v = s
	case *pb.Operand_RelativeStrengthIndex:
		if val.RelativeStrengthIndex != pb.RelativeStrengthIndex_RELATIVE_STRENGTH_INDEX_RSI {
			return nil, fmt.Errorf("unknown RelativeStrengthIndex value: %v", val.RelativeStrengthIndex)
		}
		v = "rsi"
	case *pb.Operand_Volume:
		if val.Volume != pb.Volume_VOLUME_VOLUME {
			return nil, fmt.Errorf("unknown Volume value: %v", val.Volume)
		}
		v = "volume"
	case *pb.Operand_ConstValue:
		v = val.ConstValue
	default:
		return nil, fmt.Errorf("unset operand value")
	}
	return json.Marshal(v)
}

// jsonToProtoOperand parses a JSON operand value (string or number) into a proto Operand.
func jsonToProtoOperand(raw json.RawMessage) (*pb.Operand, error) {
	// Try numeric (const value) first — JSON numbers don't coerce to string.
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return &pb.Operand{Value: &pb.Operand_ConstValue{ConstValue: f}}, nil
	}

	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("invalid operand JSON: %w", err)
	}

	switch s {
	case "open":
		return &pb.Operand{Value: &pb.Operand_Price{Price: pb.Price_PRICE_OPEN}}, nil
	case "high":
		return &pb.Operand{Value: &pb.Operand_Price{Price: pb.Price_PRICE_HIGH}}, nil
	case "low":
		return &pb.Operand{Value: &pb.Operand_Price{Price: pb.Price_PRICE_LOW}}, nil
	case "close":
		return &pb.Operand{Value: &pb.Operand_Price{Price: pb.Price_PRICE_CLOSE}}, nil
	case "middle":
		return &pb.Operand{Value: &pb.Operand_BollingerBand{BollingerBand: pb.BollingerBand_BOLLINGER_BAND_MIDDLE}}, nil
	case "lower":
		return &pb.Operand{Value: &pb.Operand_BollingerBand{BollingerBand: pb.BollingerBand_BOLLINGER_BAND_LOWER}}, nil
	case "upper":
		return &pb.Operand{Value: &pb.Operand_BollingerBand{BollingerBand: pb.BollingerBand_BOLLINGER_BAND_UPPER}}, nil
	case "sma10":
		return &pb.Operand{Value: &pb.Operand_SimpleMovingAverage{SimpleMovingAverage: pb.SimpleMovingAverage_SIMPLE_MOVING_AVERAGE_SMA10}}, nil
	case "sma50":
		return &pb.Operand{Value: &pb.Operand_SimpleMovingAverage{SimpleMovingAverage: pb.SimpleMovingAverage_SIMPLE_MOVING_AVERAGE_SMA50}}, nil
	case "ma100":
		return &pb.Operand{Value: &pb.Operand_SimpleMovingAverage{SimpleMovingAverage: pb.SimpleMovingAverage_SIMPLE_MOVING_AVERAGE_SMA100}}, nil
	case "ma200":
		return &pb.Operand{Value: &pb.Operand_SimpleMovingAverage{SimpleMovingAverage: pb.SimpleMovingAverage_SIMPLE_MOVING_AVERAGE_SMA200}}, nil
	case "rsi":
		return &pb.Operand{Value: &pb.Operand_RelativeStrengthIndex{RelativeStrengthIndex: pb.RelativeStrengthIndex_RELATIVE_STRENGTH_INDEX_RSI}}, nil
	case "volume":
		return &pb.Operand{Value: &pb.Operand_Volume{Volume: pb.Volume_VOLUME_VOLUME}}, nil
	default:
		return nil, fmt.Errorf("unknown operand value: %q", s)
	}
}

// operand1ToProto converts an oapi Alert_Operand1 union to a proto Operand.
func operand1ToProto(o oapi.Alert_Operand1) (*pb.Operand, error) {
	raw, err := o.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return jsonToProtoOperand(raw)
}

// operand2ToProto converts an oapi Alert_Operand2 union to a proto Operand.
func operand2ToProto(o oapi.Alert_Operand2) (*pb.Operand, error) {
	raw, err := o.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return jsonToProtoOperand(raw)
}

// operand1FromProto converts a proto Operand to an oapi Alert_Operand1 union.
func operand1FromProto(o *pb.Operand) (oapi.Alert_Operand1, error) {
	var result oapi.Alert_Operand1
	b, err := protoOperandToJSON(o)
	if err != nil {
		return result, err
	}
	return result, result.UnmarshalJSON(b)
}

// operand2FromProto converts a proto Operand to an oapi Alert_Operand2 union.
func operand2FromProto(o *pb.Operand) (oapi.Alert_Operand2, error) {
	var result oapi.Alert_Operand2
	b, err := protoOperandToJSON(o)
	if err != nil {
		return result, err
	}
	return result, result.UnmarshalJSON(b)
}

// validateOperandCompatibility enforces the spec rules:
//   - operand1 must be price-based or niche-based (never a constant)
//   - price-based operand1 requires operand2 to be constant or price-based
//   - niche-based operand1 requires operand2 to be constant or niche-based
func validateOperandCompatibility(op1, op2 *pb.Operand) error {
	isPriceBased := func(o *pb.Operand) bool {
		switch o.Value.(type) {
		case *pb.Operand_Price, *pb.Operand_BollingerBand, *pb.Operand_SimpleMovingAverage:
			return true
		}
		return false
	}
	isNicheBased := func(o *pb.Operand) bool {
		switch o.Value.(type) {
		case *pb.Operand_RelativeStrengthIndex, *pb.Operand_Volume:
			return true
		}
		return false
	}
	isConst := func(o *pb.Operand) bool {
		_, ok := o.Value.(*pb.Operand_ConstValue)
		return ok
	}

	switch {
	case isPriceBased(op1):
		if !isConst(op2) && !isPriceBased(op2) {
			return fmt.Errorf("price-based operand1 requires operand2 to be a constant value or price-based indicator")
		}
	case isNicheBased(op1):
		if !isConst(op2) && !isNicheBased(op2) {
			return fmt.Errorf("niche-based operand1 requires operand2 to be a constant value or niche-based indicator")
		}
	default:
		return fmt.Errorf("operand1 must be a price-based or niche-based indicator, not a constant value")
	}
	return nil
}

// alertFromProto converts a gRPC Alert to an OpenAPI Alert.
func alertFromProto(a *pb.Alert) (oapi.Alert, error) {
	op1, err := operand1FromProto(a.Operand1)
	if err != nil {
		return oapi.Alert{}, fmt.Errorf("operand1: %w", err)
	}
	op2, err := operand2FromProto(a.Operand2)
	if err != nil {
		return oapi.Alert{}, fmt.Errorf("operand2: %w", err)
	}

	out := oapi.Alert{
		Id:        utils.Ptr(a.Id),
		Timeframe: timeframeFromProto[a.Timeframe],
		Symbol:    a.Symbol,
		Operand1:  op1,
		Operand2:  op2,
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
	return out, nil
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
		alert, err := alertFromProto(a)
		if err != nil {
			logger.Error("failed to convert alert from proto", zap.Error(err))
			c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
				Code:    "InternalServerError",
				Message: "failed to process alert data",
			})
			return
		}
		out = append(out, alert)
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

	op1, err := operand1ToProto(req.Operand1)
	if err != nil {
		logger.Error("invalid operand1", zap.Error(err))
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "invalid operand1: " + err.Error(),
		})
		return
	}

	op2, err := operand2ToProto(req.Operand2)
	if err != nil {
		logger.Error("invalid operand2", zap.Error(err))
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "invalid operand2: " + err.Error(),
		})
		return
	}

	if err := validateOperandCompatibility(op1, op2); err != nil {
		logger.Error("incompatible operands", zap.Error(err))
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: err.Error(),
		})
		return
	}

	resp, err := s.strategyClient.CreateAlert(c, &pb.CreateAlertRequest{
		OfAccountId: accountId,
		Timeframe:   timeframeToProto[req.Timeframe],
		Symbol:      req.Symbol,
		Operand1:    op1,
		Operator:    operatorToProto[req.Operator],
		Trigger:     triggerToProto[req.Trigger],
		Exp:         req.Exp,
		Message:     req.Message,
		Operand2:    op2,
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

	alert, err := alertFromProto(resp.Alert)
	if err != nil {
		logger.Error("failed to convert created alert from proto", zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: "failed to process alert data",
		})
		return
	}

	c.JSON(http.StatusCreated, alert)
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

	alert, err := alertFromProto(resp.Alert)
	if err != nil {
		logger.Error("failed to convert alert from proto", zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: "failed to process alert data",
		})
		return
	}

	c.JSON(http.StatusOK, alert)
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

	op1, err := operand1ToProto(req.Operand1)
	if err != nil {
		logger.Error("invalid operand1", zap.Error(err))
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "invalid operand1: " + err.Error(),
		})
		return
	}

	op2, err := operand2ToProto(req.Operand2)
	if err != nil {
		logger.Error("invalid operand2", zap.Error(err))
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: "invalid operand2: " + err.Error(),
		})
		return
	}

	if err := validateOperandCompatibility(op1, op2); err != nil {
		logger.Error("incompatible operands", zap.Error(err))
		c.JSON(http.StatusBadRequest, oapi.BadRequest{
			Code:    "BadRequest",
			Message: err.Error(),
		})
		return
	}

	resp, err := s.strategyClient.UpdateAlert(c, &pb.UpdateAlertRequest{
		OfAccountId: accountId,
		AlertId:     alertId,
		Timeframe:   timeframeToProto[req.Timeframe],
		Symbol:      req.Symbol,
		Operand1:    op1,
		Operator:    operatorToProto[req.Operator],
		Trigger:     triggerToProto[req.Trigger],
		Exp:         req.Exp,
		Message:     req.Message,
		Operand2:    op2,
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

	alert, err := alertFromProto(resp.Alert)
	if err != nil {
		logger.Error("failed to convert updated alert from proto", zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi.InternalServerError{
			Code:    "InternalServerError",
			Message: "failed to process alert data",
		})
		return
	}

	c.JSON(http.StatusOK, alert)
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
