package logic_http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "github.com/Fiagram/standalone/internal/generated/grpc/strategy"
	oapi "github.com/Fiagram/standalone/internal/generated/openapi"
	logic_http "github.com/Fiagram/standalone/internal/logic/http"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ---------------------------------------------------------------------------
// Mock: dao_strategy.Client
// ---------------------------------------------------------------------------

type mockStrategyClient struct {
	createAlertFn func(ctx context.Context, in *pb.CreateAlertRequest, opts ...grpc.CallOption) (*pb.CreateAlertResponse, error)
	getAlertsFn   func(ctx context.Context, in *pb.GetAlertsRequest, opts ...grpc.CallOption) (*pb.GetAlertsResponse, error)
	getAlertFn    func(ctx context.Context, in *pb.GetAlertRequest, opts ...grpc.CallOption) (*pb.GetAlertResponse, error)
	updateAlertFn func(ctx context.Context, in *pb.UpdateAlertRequest, opts ...grpc.CallOption) (*pb.UpdateAlertResponse, error)
	deleteAlertFn func(ctx context.Context, in *pb.DeleteAlertRequest, opts ...grpc.CallOption) (*pb.DeleteAlertResponse, error)
}

func (m *mockStrategyClient) CreateAlert(ctx context.Context, in *pb.CreateAlertRequest, opts ...grpc.CallOption) (*pb.CreateAlertResponse, error) {
	if m.createAlertFn != nil {
		return m.createAlertFn(ctx, in, opts...)
	}
	return &pb.CreateAlertResponse{}, nil
}

func (m *mockStrategyClient) GetAlerts(ctx context.Context, in *pb.GetAlertsRequest, opts ...grpc.CallOption) (*pb.GetAlertsResponse, error) {
	if m.getAlertsFn != nil {
		return m.getAlertsFn(ctx, in, opts...)
	}
	return &pb.GetAlertsResponse{}, nil
}

func (m *mockStrategyClient) GetAlert(ctx context.Context, in *pb.GetAlertRequest, opts ...grpc.CallOption) (*pb.GetAlertResponse, error) {
	if m.getAlertFn != nil {
		return m.getAlertFn(ctx, in, opts...)
	}
	return &pb.GetAlertResponse{}, nil
}

func (m *mockStrategyClient) UpdateAlert(ctx context.Context, in *pb.UpdateAlertRequest, opts ...grpc.CallOption) (*pb.UpdateAlertResponse, error) {
	if m.updateAlertFn != nil {
		return m.updateAlertFn(ctx, in, opts...)
	}
	return &pb.UpdateAlertResponse{}, nil
}

func (m *mockStrategyClient) DeleteAlert(ctx context.Context, in *pb.DeleteAlertRequest, opts ...grpc.CallOption) (*pb.DeleteAlertResponse, error) {
	if m.deleteAlertFn != nil {
		return m.deleteAlertFn(ctx, in, opts...)
	}
	return &pb.DeleteAlertResponse{}, nil
}

func (m *mockStrategyClient) Close() error {
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestStrategyLogic(client *mockStrategyClient) logic_http.StrategyLogic {
	return logic_http.NewStrategyLogic(client, zap.NewNop())
}

func sampleProtoAlert() *pb.Alert {
	msg := "test alert message"
	return &pb.Alert{
		Id:        "1",
		Timeframe: pb.Timeframe_TIMEFRAME_D1,
		Symbol:    "AAPL",
		Operand1:  &pb.Operand{Value: &pb.Operand_Price{Price: pb.Price_PRICE_CLOSE}},
		Operator:  pb.Operator_OPERATOR_GREATER_THAN,
		Trigger:   pb.Trigger_TRIGGER_ONCE,
		Exp:       1700000000,
		Message:   &msg,
		Operand2:  &pb.Operand{Value: &pb.Operand_ConstValue{ConstValue: 150.0}},
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
	}
}

// ---------------------------------------------------------------------------
// Tests: GetStrategyAlerts
// ---------------------------------------------------------------------------

func TestGetStrategyAlerts_Success_Defaults(t *testing.T) {
	mock := &mockStrategyClient{
		getAlertsFn: func(_ context.Context, in *pb.GetAlertsRequest, _ ...grpc.CallOption) (*pb.GetAlertsResponse, error) {
			require.Equal(t, uint64(42), in.OfAccountId)
			require.Equal(t, uint32(20), in.Limit)
			require.Equal(t, uint32(0), in.Offset)
			return &pb.GetAlertsResponse{
				Alerts: []*pb.Alert{sampleProtoAlert()},
			}, nil
		},
	}
	sl := newTestStrategyLogic(mock)

	c, w := newGinContext(http.MethodGet, "/strategy/alerts", nil)
	setAccountId(c, 42)
	sl.GetStrategyAlerts(c, oapi.GetStrategyAlertsParams{})

	require.Equal(t, http.StatusOK, w.Code)

	var resp []oapi.Alert
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp, 1)
	require.Equal(t, "AAPL", resp[0].Symbol)
	require.Equal(t, oapi.D1, resp[0].Timeframe)
	require.Equal(t, oapi.GreaterThan, resp[0].Operator)
	require.Equal(t, oapi.Once, resp[0].Trigger)
	require.Equal(t, int64(1700000000), resp[0].Exp)
	require.NotNil(t, resp[0].Message)
	require.Equal(t, "test alert message", *resp[0].Message)
	require.NotNil(t, resp[0].CreatedAt)
	require.NotNil(t, resp[0].UpdatedAt)

	// operand1 should be "close" (price-based)
	op1Raw, err := resp[0].Operand1.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, `"close"`, string(op1Raw))

	// operand2 should be 150.0 (const value)
	op2Raw, err := resp[0].Operand2.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, `150`, string(op2Raw))
}

func TestGetStrategyAlerts_CustomLimitOffset(t *testing.T) {
	limit := 5
	offset := 10
	mock := &mockStrategyClient{
		getAlertsFn: func(_ context.Context, in *pb.GetAlertsRequest, _ ...grpc.CallOption) (*pb.GetAlertsResponse, error) {
			require.Equal(t, uint32(5), in.Limit)
			require.Equal(t, uint32(10), in.Offset)
			return &pb.GetAlertsResponse{}, nil
		},
	}
	sl := newTestStrategyLogic(mock)

	c, w := newGinContext(http.MethodGet, "/strategy/alerts?limit=5&offset=10", nil)
	setAccountId(c, 1)
	sl.GetStrategyAlerts(c, oapi.GetStrategyAlertsParams{
		Limit:  &limit,
		Offset: &offset,
	})

	require.Equal(t, http.StatusOK, w.Code)
}

func TestGetStrategyAlerts_Empty(t *testing.T) {
	mock := &mockStrategyClient{
		getAlertsFn: func(_ context.Context, _ *pb.GetAlertsRequest, _ ...grpc.CallOption) (*pb.GetAlertsResponse, error) {
			return &pb.GetAlertsResponse{Alerts: []*pb.Alert{}}, nil
		},
	}
	sl := newTestStrategyLogic(mock)

	c, w := newGinContext(http.MethodGet, "/strategy/alerts", nil)
	setAccountId(c, 1)
	sl.GetStrategyAlerts(c, oapi.GetStrategyAlertsParams{})

	require.Equal(t, http.StatusOK, w.Code)

	var resp []oapi.Alert
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp, 0)
}

func TestGetStrategyAlerts_NoAccountId(t *testing.T) {
	sl := newTestStrategyLogic(&mockStrategyClient{})

	c, w := newGinContext(http.MethodGet, "/strategy/alerts", nil)
	sl.GetStrategyAlerts(c, oapi.GetStrategyAlertsParams{})

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetStrategyAlerts_GrpcError(t *testing.T) {
	mock := &mockStrategyClient{
		getAlertsFn: func(_ context.Context, _ *pb.GetAlertsRequest, _ ...grpc.CallOption) (*pb.GetAlertsResponse, error) {
			return nil, errors.New("grpc unavailable")
		},
	}
	sl := newTestStrategyLogic(mock)

	c, w := newGinContext(http.MethodGet, "/strategy/alerts", nil)
	setAccountId(c, 1)
	sl.GetStrategyAlerts(c, oapi.GetStrategyAlertsParams{})

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: CreateStrategyAlert
// ---------------------------------------------------------------------------

func TestCreateStrategyAlert_Success(t *testing.T) {
	mock := &mockStrategyClient{
		createAlertFn: func(_ context.Context, in *pb.CreateAlertRequest, _ ...grpc.CallOption) (*pb.CreateAlertResponse, error) {
			require.Equal(t, uint64(7), in.OfAccountId)
			require.Equal(t, pb.Timeframe_TIMEFRAME_W1, in.Timeframe)
			require.Equal(t, "TSLA", in.Symbol)
			require.NotNil(t, in.Operand1)
			require.NotNil(t, in.Operand1.GetRelativeStrengthIndex())
			require.Equal(t, pb.RelativeStrengthIndex_RELATIVE_STRENGTH_INDEX_RSI, in.Operand1.GetRelativeStrengthIndex())
			require.Equal(t, pb.Operator_OPERATOR_LESS_THAN, in.Operator)
			require.Equal(t, pb.Trigger_TRIGGER_EVERY, in.Trigger)
			require.Equal(t, int64(1800000000), in.Exp)
			require.NotNil(t, in.Message)
			require.Equal(t, "buy signal", *in.Message)
			require.NotNil(t, in.Operand2)
			require.Equal(t, 70.0, in.Operand2.GetConstValue())

			msg := "buy signal"
			return &pb.CreateAlertResponse{
				Alert: &pb.Alert{
					Id:        "100",
					Timeframe: in.Timeframe,
					Symbol:    in.Symbol,
					Operand1:  in.Operand1,
					Operator:  in.Operator,
					Trigger:   in.Trigger,
					Exp:       in.Exp,
					Message:   &msg,
					Operand2:  in.Operand2,
					CreatedAt: timestamppb.Now(),
				},
			}, nil
		},
	}
	sl := newTestStrategyLogic(mock)

	msg := "buy signal"
	var op1 oapi.Alert_Operand1
	require.NoError(t, op1.UnmarshalJSON([]byte(`"rsi"`)))
	var op2 oapi.Alert_Operand2
	require.NoError(t, op2.UnmarshalJSON([]byte(`70`)))
	body := oapi.Alert{
		Timeframe: oapi.W1,
		Symbol:    "TSLA",
		Operand1:  op1,
		Operator:  oapi.LessThan,
		Trigger:   oapi.Every,
		Exp:       1800000000,
		Message:   &msg,
		Operand2:  op2,
	}
	c, w := newGinContext(http.MethodPost, "/strategy/alerts", body)
	setAccountId(c, 7)
	sl.CreateStrategyAlert(c)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp oapi.Alert
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.NotNil(t, resp.Id)
	require.Equal(t, "100", *resp.Id)
	require.Equal(t, oapi.W1, resp.Timeframe)
	require.Equal(t, "TSLA", resp.Symbol)
	require.Equal(t, oapi.LessThan, resp.Operator)
	require.Equal(t, oapi.Every, resp.Trigger)
}

func TestCreateStrategyAlert_NoAccountId(t *testing.T) {
	sl := newTestStrategyLogic(&mockStrategyClient{})

	var op1 oapi.Alert_Operand1
	require.NoError(t, op1.UnmarshalJSON([]byte(`"close"`)))
	var op2 oapi.Alert_Operand2
	require.NoError(t, op2.UnmarshalJSON([]byte(`100`)))
	body := oapi.Alert{
		Timeframe: oapi.D1,
		Symbol:    "GOOG",
		Operand1:  op1,
		Operator:  oapi.GreaterThan,
		Trigger:   oapi.Once,
		Exp:       1700000000,
		Operand2:  op2,
	}
	c, w := newGinContext(http.MethodPost, "/strategy/alerts", body)
	sl.CreateStrategyAlert(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateStrategyAlert_InvalidBody(t *testing.T) {
	sl := newTestStrategyLogic(&mockStrategyClient{})

	w, c := newInvalidBodyContext(http.MethodPost, "/strategy/alerts")
	setAccountId(c, 1)
	sl.CreateStrategyAlert(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateStrategyAlert_GrpcError(t *testing.T) {
	mock := &mockStrategyClient{
		createAlertFn: func(_ context.Context, _ *pb.CreateAlertRequest, _ ...grpc.CallOption) (*pb.CreateAlertResponse, error) {
			return nil, errors.New("strategy service down")
		},
	}
	sl := newTestStrategyLogic(mock)

	var op1 oapi.Alert_Operand1
	require.NoError(t, op1.UnmarshalJSON([]byte(`"sma50"`)))
	var op2 oapi.Alert_Operand2
	require.NoError(t, op2.UnmarshalJSON([]byte(`200`)))
	body := oapi.Alert{
		Timeframe: oapi.D1,
		Symbol:    "MSFT",
		Operand1:  op1,
		Operator:  oapi.Crossing,
		Trigger:   oapi.Once,
		Exp:       1700000000,
		Operand2:  op2,
	}
	c, w := newGinContext(http.MethodPost, "/strategy/alerts", body)
	setAccountId(c, 1)
	sl.CreateStrategyAlert(c)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateStrategyAlert_IncompatibleOperands(t *testing.T) {
	sl := newTestStrategyLogic(&mockStrategyClient{})

	// price-based operand1 + niche-based operand2 → incompatible
	var op1 oapi.Alert_Operand1
	require.NoError(t, op1.UnmarshalJSON([]byte(`"close"`)))
	var op2 oapi.Alert_Operand2
	require.NoError(t, op2.UnmarshalJSON([]byte(`"rsi"`)))
	body := oapi.Alert{
		Timeframe: oapi.D1,
		Symbol:    "AAPL",
		Operand1:  op1,
		Operator:  oapi.GreaterThan,
		Trigger:   oapi.Once,
		Exp:       1700000000,
		Operand2:  op2,
	}
	c, w := newGinContext(http.MethodPost, "/strategy/alerts", body)
	setAccountId(c, 1)
	sl.CreateStrategyAlert(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GetStrategyAlert
// ---------------------------------------------------------------------------

func TestGetStrategyAlert_Success(t *testing.T) {
	mock := &mockStrategyClient{
		getAlertFn: func(_ context.Context, in *pb.GetAlertRequest, _ ...grpc.CallOption) (*pb.GetAlertResponse, error) {
			require.Equal(t, uint64(5), in.OfAccountId)
			require.Equal(t, "99", in.AlertId)
			return &pb.GetAlertResponse{
				Alert: sampleProtoAlert(),
			}, nil
		},
	}
	sl := newTestStrategyLogic(mock)

	c, w := newGinContext(http.MethodGet, "/strategy/alerts/99", nil)
	setAccountId(c, 5)
	sl.GetStrategyAlert(c, "99")

	require.Equal(t, http.StatusOK, w.Code)

	var resp oapi.Alert
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.NotNil(t, resp.Id)
	require.Equal(t, "1", *resp.Id)
	require.Equal(t, "AAPL", resp.Symbol)
}

func TestGetStrategyAlert_NoAccountId(t *testing.T) {
	sl := newTestStrategyLogic(&mockStrategyClient{})

	c, w := newGinContext(http.MethodGet, "/strategy/alerts/1", nil)
	sl.GetStrategyAlert(c, "1")

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetStrategyAlert_NotFound(t *testing.T) {
	mock := &mockStrategyClient{
		getAlertFn: func(_ context.Context, _ *pb.GetAlertRequest, _ ...grpc.CallOption) (*pb.GetAlertResponse, error) {
			return nil, errors.New("not found")
		},
	}
	sl := newTestStrategyLogic(mock)

	c, w := newGinContext(http.MethodGet, "/strategy/alerts/999", nil)
	setAccountId(c, 1)
	sl.GetStrategyAlert(c, "999")

	require.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: UpdateStrategyAlert
// ---------------------------------------------------------------------------

func TestUpdateStrategyAlert_Success(t *testing.T) {
	mock := &mockStrategyClient{
		updateAlertFn: func(_ context.Context, in *pb.UpdateAlertRequest, _ ...grpc.CallOption) (*pb.UpdateAlertResponse, error) {
			require.Equal(t, uint64(3), in.OfAccountId)
			require.Equal(t, "50", in.AlertId)
			require.Equal(t, pb.Timeframe_TIMEFRAME_M1, in.Timeframe)
			require.Equal(t, "AMZN", in.Symbol)
			require.NotNil(t, in.Operand1)
			require.Equal(t, pb.BollingerBand_BOLLINGER_BAND_LOWER, in.Operand1.GetBollingerBand())
			require.Equal(t, pb.Operator_OPERATOR_CROSSING_UP, in.Operator)
			require.Equal(t, pb.Trigger_TRIGGER_ONCE, in.Trigger)

			return &pb.UpdateAlertResponse{
				Alert: &pb.Alert{
					Id:        "50",
					Timeframe: in.Timeframe,
					Symbol:    in.Symbol,
					Operand1:  in.Operand1,
					Operator:  in.Operator,
					Trigger:   in.Trigger,
					Exp:       in.Exp,
					Message:   in.Message,
					Operand2:  in.Operand2,
					UpdatedAt: timestamppb.Now(),
				},
			}, nil
		},
	}
	sl := newTestStrategyLogic(mock)

	msg := "updated alert"
	var op1 oapi.Alert_Operand1
	require.NoError(t, op1.UnmarshalJSON([]byte(`"lower"`)))
	var op2 oapi.Alert_Operand2
	require.NoError(t, op2.UnmarshalJSON([]byte(`"close"`)))
	body := oapi.Alert{
		Timeframe: oapi.M1,
		Symbol:    "AMZN",
		Operand1:  op1,
		Operator:  oapi.CrossingUp,
		Trigger:   oapi.Once,
		Exp:       1900000000,
		Message:   &msg,
		Operand2:  op2,
	}
	c, w := newGinContext(http.MethodPut, "/strategy/alerts/50", body)
	setAccountId(c, 3)
	sl.UpdateStrategyAlert(c, "50")

	require.Equal(t, http.StatusOK, w.Code)

	var resp oapi.Alert
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.NotNil(t, resp.Id)
	require.Equal(t, "50", *resp.Id)
	require.Equal(t, oapi.M1, resp.Timeframe)
	require.Equal(t, "AMZN", resp.Symbol)
	require.Equal(t, oapi.CrossingUp, resp.Operator)
	require.NotNil(t, resp.UpdatedAt)
}

func TestUpdateStrategyAlert_NoAccountId(t *testing.T) {
	sl := newTestStrategyLogic(&mockStrategyClient{})

	var op1 oapi.Alert_Operand1
	require.NoError(t, op1.UnmarshalJSON([]byte(`"close"`)))
	var op2 oapi.Alert_Operand2
	require.NoError(t, op2.UnmarshalJSON([]byte(`50`)))
	body := oapi.Alert{
		Timeframe: oapi.D1,
		Symbol:    "X",
		Operand1:  op1,
		Operator:  oapi.GreaterThan,
		Trigger:   oapi.Once,
		Exp:       1700000000,
		Operand2:  op2,
	}
	c, w := newGinContext(http.MethodPut, "/strategy/alerts/1", body)
	sl.UpdateStrategyAlert(c, "1")

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateStrategyAlert_InvalidBody(t *testing.T) {
	sl := newTestStrategyLogic(&mockStrategyClient{})

	w, c := newInvalidBodyContext(http.MethodPut, "/strategy/alerts/1")
	setAccountId(c, 1)
	sl.UpdateStrategyAlert(c, "1")

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateStrategyAlert_GrpcError(t *testing.T) {
	mock := &mockStrategyClient{
		updateAlertFn: func(_ context.Context, _ *pb.UpdateAlertRequest, _ ...grpc.CallOption) (*pb.UpdateAlertResponse, error) {
			return nil, errors.New("update failed")
		},
	}
	sl := newTestStrategyLogic(mock)

	var op1 oapi.Alert_Operand1
	require.NoError(t, op1.UnmarshalJSON([]byte(`"ma200"`)))
	var op2 oapi.Alert_Operand2
	require.NoError(t, op2.UnmarshalJSON([]byte(`"close"`)))
	body := oapi.Alert{
		Timeframe: oapi.D1,
		Symbol:    "FB",
		Operand1:  op1,
		Operator:  oapi.CrossingDown,
		Trigger:   oapi.Every,
		Exp:       1700000000,
		Operand2:  op2,
	}
	c, w := newGinContext(http.MethodPut, "/strategy/alerts/1", body)
	setAccountId(c, 1)
	sl.UpdateStrategyAlert(c, "1")

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateStrategyAlert_IncompatibleOperands(t *testing.T) {
	sl := newTestStrategyLogic(&mockStrategyClient{})

	// niche-based operand1 + price-based operand2 → incompatible
	var op1 oapi.Alert_Operand1
	require.NoError(t, op1.UnmarshalJSON([]byte(`"rsi"`)))
	var op2 oapi.Alert_Operand2
	require.NoError(t, op2.UnmarshalJSON([]byte(`"sma50"`)))
	body := oapi.Alert{
		Timeframe: oapi.D1,
		Symbol:    "AAPL",
		Operand1:  op1,
		Operator:  oapi.GreaterThan,
		Trigger:   oapi.Once,
		Exp:       1700000000,
		Operand2:  op2,
	}
	c, w := newGinContext(http.MethodPut, "/strategy/alerts/1", body)
	setAccountId(c, 1)
	sl.UpdateStrategyAlert(c, "1")

	require.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: DeleteStrategyAlert
// ---------------------------------------------------------------------------

func TestDeleteStrategyAlert_Success(t *testing.T) {
	mock := &mockStrategyClient{
		deleteAlertFn: func(_ context.Context, in *pb.DeleteAlertRequest, _ ...grpc.CallOption) (*pb.DeleteAlertResponse, error) {
			require.Equal(t, uint64(8), in.OfAccountId)
			require.Equal(t, "77", in.AlertId)
			return &pb.DeleteAlertResponse{
				OfAccountId: in.OfAccountId,
				AlertId:     in.AlertId,
			}, nil
		},
	}
	sl := newTestStrategyLogic(mock)

	c, _ := newGinContext(http.MethodDelete, "/strategy/alerts/77", nil)
	setAccountId(c, 8)
	sl.DeleteStrategyAlert(c, "77")

	// c.Status() sets gin's internal writer status but does not flush to
	// httptest.ResponseRecorder when no body is written, so check the writer.
	require.Equal(t, http.StatusNoContent, c.Writer.Status())
}

func TestDeleteStrategyAlert_NoAccountId(t *testing.T) {
	sl := newTestStrategyLogic(&mockStrategyClient{})

	c, w := newGinContext(http.MethodDelete, "/strategy/alerts/1", nil)
	sl.DeleteStrategyAlert(c, "1")

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteStrategyAlert_GrpcError(t *testing.T) {
	mock := &mockStrategyClient{
		deleteAlertFn: func(_ context.Context, _ *pb.DeleteAlertRequest, _ ...grpc.CallOption) (*pb.DeleteAlertResponse, error) {
			return nil, errors.New("delete failed")
		},
	}
	sl := newTestStrategyLogic(mock)

	c, w := newGinContext(http.MethodDelete, "/strategy/alerts/1", nil)
	setAccountId(c, 1)
	sl.DeleteStrategyAlert(c, "1")

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Helper: invalid body context
// ---------------------------------------------------------------------------

func newInvalidBodyContext(method, path string) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path,
		bytes.NewReader([]byte(`not json`)))
	c.Request.Header.Set("Content-Type", "application/json")
	return w, c
}
