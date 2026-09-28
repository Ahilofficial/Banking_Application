package client

import (
	"context"
	"fmt"
	"time"

	pb "banking-microservices/pkg/proto/notification"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type NotificationGRPCClient struct {
	addr   string
	client pb.NotificationServiceClient
	conn   *grpc.ClientConn
}

// NewNotificationGRPCClient creates a client connected to the Notification Service via gRPC
func NewNotificationGRPCClient(addr string) (*NotificationGRPCClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial notification gRPC at %s: %w", addr, err)
	}

	return &NotificationGRPCClient{
		addr:   addr,
		client: pb.NewNotificationServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *NotificationGRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *NotificationGRPCClient) SendEmail(ctx context.Context, req *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.SendEmail(ctx, req)
}

func (c *NotificationGRPCClient) SendSMS(ctx context.Context, req *pb.SendSMSRequest) (*pb.SendSMSResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.SendSMS(ctx, req)
}

func (c *NotificationGRPCClient) SendNotification(ctx context.Context, req *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.SendNotification(ctx, req)
}

func (c *NotificationGRPCClient) GetNotifications(ctx context.Context, req *pb.GetNotificationsRequest) (*pb.GetNotificationsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.GetNotifications(ctx, req)
}

func (c *NotificationGRPCClient) GetNotificationStatus(ctx context.Context, req *pb.GetNotificationStatusRequest) (*pb.NotificationStatusResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.GetNotificationStatus(ctx, req)
}
