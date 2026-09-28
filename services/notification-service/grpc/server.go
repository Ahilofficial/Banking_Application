package grpc

import (
	"context"

	pb "banking-microservices/pkg/proto/notification"
	"banking-microservices/services/notification-service/service"

	"github.com/google/wire"
	"google.golang.org/grpc"
)

// ProviderSet provides gRPC Server dependencies for Wire
var ProviderSet = wire.NewSet(NewNotificationGRPCServer)

type NotificationGRPCServer struct {
	pb.UnimplementedNotificationServiceServer
	svc *service.NotificationService
}

func NewNotificationGRPCServer(svc *service.NotificationService) *NotificationGRPCServer {
	return &NotificationGRPCServer{svc: svc}
}

func (s *NotificationGRPCServer) Register(server *grpc.Server) {
	pb.RegisterNotificationServiceServer(server, s)
}

func (s *NotificationGRPCServer) SendEmail(ctx context.Context, req *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	return s.svc.SendEmail(ctx, req)
}

func (s *NotificationGRPCServer) SendSMS(ctx context.Context, req *pb.SendSMSRequest) (*pb.SendSMSResponse, error) {
	return s.svc.SendSMS(ctx, req)
}

func (s *NotificationGRPCServer) SendNotification(ctx context.Context, req *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	return s.svc.SendNotification(ctx, req)
}

func (s *NotificationGRPCServer) GetNotifications(ctx context.Context, req *pb.GetNotificationsRequest) (*pb.GetNotificationsResponse, error) {
	return s.svc.GetNotifications(ctx, req)
}

func (s *NotificationGRPCServer) GetNotificationStatus(ctx context.Context, req *pb.GetNotificationStatusRequest) (*pb.NotificationStatusResponse, error) {
	return s.svc.GetNotificationStatus(ctx, req)
}
