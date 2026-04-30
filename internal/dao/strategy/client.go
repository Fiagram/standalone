package dao_strategy

import (
	"context"
	"fmt"

	"github.com/Fiagram/standalone/internal/configs"
	pb "github.com/Fiagram/standalone/internal/generated/grpc/strategy"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client interface {
	pb.StrategyClient
	Close() error
}

type client struct {
	stub pb.StrategyClient
	conn *grpc.ClientConn
}

func NewClient(
	config configs.Strategy,
	logger *zap.Logger,
) (Client, error) {
	logger.With(zap.Any("grpc_strategy_config", config))

	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", config.Address, config.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("failed to init grpc connection", zap.Error(err))
		return nil, fmt.Errorf("failed to init grpc connection")
	}
	stub := pb.NewStrategyClient(conn)

	return &client{
		stub: stub,
		conn: conn,
	}, nil
}

func (c *client) CreateAlert(
	ctx context.Context,
	in *pb.CreateAlertRequest,
	opts ...grpc.CallOption,
) (*pb.CreateAlertResponse, error) {
	return c.stub.CreateAlert(ctx, in, opts...)
}

func (c *client) GetAlerts(
	ctx context.Context,
	in *pb.GetAlertsRequest,
	opts ...grpc.CallOption,
) (*pb.GetAlertsResponse, error) {
	return c.stub.GetAlerts(ctx, in, opts...)
}

func (c *client) GetAlert(
	ctx context.Context,
	in *pb.GetAlertRequest,
	opts ...grpc.CallOption,
) (*pb.GetAlertResponse, error) {
	return c.stub.GetAlert(ctx, in, opts...)
}

func (c *client) UpdateAlert(
	ctx context.Context,
	in *pb.UpdateAlertRequest,
	opts ...grpc.CallOption,
) (*pb.UpdateAlertResponse, error) {
	return c.stub.UpdateAlert(ctx, in, opts...)
}

func (c *client) DeleteAlert(
	ctx context.Context,
	in *pb.DeleteAlertRequest,
	opts ...grpc.CallOption,
) (*pb.DeleteAlertResponse, error) {
	return c.stub.DeleteAlert(ctx, in, opts...)
}

func (c *client) Close() error {
	return c.conn.Close()
}
