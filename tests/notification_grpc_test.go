package tests

import (
	"context"
	"net"
	"os"
	"testing"

	"banking-microservices/pkg/common/database"
	pb "banking-microservices/pkg/proto/notification"
	grpcService "banking-microservices/services/notification-service/grpc"
	"banking-microservices/services/notification-service/models"
	"banking-microservices/services/notification-service/repository"
	"banking-microservices/services/notification-service/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"gorm.io/gorm"
)

const bufSize = 1024 * 1024

func setupNotificationGRPCTestServer(t *testing.T) (pb.NotificationServiceClient, *gorm.DB, func()) {
	_ = os.Remove("test_notification.db")
	db, err := database.Connect("sqlite", "test_notification.db")
	if err != nil {
		t.Fatalf("Failed to initialize test db: %v", err)
	}

	repo := repository.NewNotificationRepository(db)
	if err := repo.AutoMigrate(); err != nil {
		t.Fatalf("Failed to auto-migrate notification repo: %v", err)
	}

	svc := service.NewNotificationService(repo, nil)
	grpcServerImpl := grpcService.NewNotificationGRPCServer(svc)

	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	grpcServerImpl.Register(s)

	go func() {
		if err := s.Serve(lis); err != nil {
			// server stopped
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	client := pb.NewNotificationServiceClient(conn)

	cleanup := func() {
		_ = conn.Close()
		s.Stop()
		_ = lis.Close()
		_ = os.Remove("test_notification.db")
	}

	return client, db, cleanup
}

func TestNotificationGRPC_SendEmail(t *testing.T) {
	client, db, cleanup := setupNotificationGRPCTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// 1. Send Email over gRPC
	req := &pb.SendEmailRequest{
		UserId:   101,
		To:       "customer@example.com",
		Subject:  "Account Balance Alert",
		Body:     "Your current balance is $5,000.00",
		HtmlBody: "<h3>Your current balance is $5,000.00</h3>",
		Metadata: map[string]string{
			"alert_type": "BALANCE_UPDATE",
		},
	}

	resp, err := client.SendEmail(ctx, req)
	if err != nil {
		t.Fatalf("SendEmail failed over gRPC: %v", err)
	}

	if resp.NotificationId == "" {
		t.Fatal("Expected non-empty NotificationId")
	}
	if resp.Status != pb.NotificationStatus_SENT {
		t.Fatalf("Expected status SENT, got %v", resp.Status)
	}
	if resp.Message != "Email sent successfully" {
		t.Fatalf("Unexpected message: %s", resp.Message)
	}

	// Verify database persistence
	var n models.Notification
	if err := db.First(&n, "id = ?", resp.NotificationId).Error; err != nil {
		t.Fatalf("Failed to find notification record in DB: %v", err)
	}
	if n.Recipient != "customer@example.com" || n.Subject != "Account Balance Alert" {
		t.Fatalf("DB record mismatch: %+v", n)
	}
}

func TestNotificationGRPC_SendSMS(t *testing.T) {
	client, db, cleanup := setupNotificationGRPCTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// 2. Send SMS over gRPC
	req := &pb.SendSMSRequest{
		UserId:      102,
		PhoneNumber: "+1234567890",
		Message:     "Your OTP code is 492019 for login verification.",
		Metadata: map[string]string{
			"channel": "TRANSACTIONAL_SMS",
		},
	}

	resp, err := client.SendSMS(ctx, req)
	if err != nil {
		t.Fatalf("SendSMS failed over gRPC: %v", err)
	}

	if resp.NotificationId == "" {
		t.Fatal("Expected non-empty NotificationId")
	}
	if resp.Status != pb.NotificationStatus_SENT {
		t.Fatalf("Expected status SENT, got %v", resp.Status)
	}

	// Verify database persistence
	var n models.Notification
	if err := db.First(&n, "id = ?", resp.NotificationId).Error; err != nil {
		t.Fatalf("Failed to find SMS in DB: %v", err)
	}
	if n.Type != "SMS" || n.Recipient != "+1234567890" {
		t.Fatalf("DB record mismatch: %+v", n)
	}
}

func TestNotificationGRPC_SendNotification(t *testing.T) {
	client, db, cleanup := setupNotificationGRPCTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// 3. Send Push Notification over gRPC
	req := &pb.SendNotificationRequest{
		UserId:    103,
		Type:      pb.NotificationType_PUSH,
		Recipient: "device-token-fcm-xyz-987",
		Title:     "Money Received!",
		Content:   "You received $250.00 from John Doe.",
		Metadata: map[string]string{
			"deep_link": "/transactions/view/999",
		},
	}

	resp, err := client.SendNotification(ctx, req)
	if err != nil {
		t.Fatalf("SendNotification failed over gRPC: %v", err)
	}

	if resp.NotificationId == "" {
		t.Fatal("Expected non-empty NotificationId")
	}

	// Verify database persistence
	var n models.Notification
	if err := db.First(&n, "id = ?", resp.NotificationId).Error; err != nil {
		t.Fatalf("Failed to find notification in DB: %v", err)
	}
	if n.Type != "PUSH" || n.Subject != "Money Received!" {
		t.Fatalf("DB record mismatch: %+v", n)
	}
}

func TestNotificationGRPC_GetNotificationsAndStatus(t *testing.T) {
	client, _, cleanup := setupNotificationGRPCTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Seed multiple notifications for user 200
	userID := uint32(200)
	emailResp, err := client.SendEmail(ctx, &pb.SendEmailRequest{
		UserId:  userID,
		To:      "user200@bank.com",
		Subject: "Welcome to Banking",
		Body:    "Welcome aboard!",
	})
	if err != nil {
		t.Fatalf("Seed email failed: %v", err)
	}

	_, err = client.SendSMS(ctx, &pb.SendSMSRequest{
		UserId:      userID,
		PhoneNumber: "+9876543210",
		Message:     "Account created",
	})
	if err != nil {
		t.Fatalf("Seed SMS failed: %v", err)
	}

	_, err = client.SendNotification(ctx, &pb.SendNotificationRequest{
		UserId:    userID,
		Type:      pb.NotificationType_IN_APP,
		Recipient: "in-app-user-200",
		Title:     "Security Alert",
		Content:   "New login from Chrome Windows",
	})
	if err != nil {
		t.Fatalf("Seed notification failed: %v", err)
	}

	// 4. Retrieve notifications for user 200 via gRPC
	getResp, err := client.GetNotifications(ctx, &pb.GetNotificationsRequest{
		UserId: userID,
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("GetNotifications failed over gRPC: %v", err)
	}

	if getResp.Total != 3 {
		t.Fatalf("Expected 3 notifications, got %d", getResp.Total)
	}
	if len(getResp.Notifications) != 3 {
		t.Fatalf("Expected 3 items in response, got %d", len(getResp.Notifications))
	}

	// 5. Query status of specific notification via gRPC
	statusResp, err := client.GetNotificationStatus(ctx, &pb.GetNotificationStatusRequest{
		NotificationId: emailResp.NotificationId,
	})
	if err != nil {
		t.Fatalf("GetNotificationStatus failed over gRPC: %v", err)
	}

	if statusResp.NotificationId != emailResp.NotificationId {
		t.Fatalf("Expected ID %s, got %s", emailResp.NotificationId, statusResp.NotificationId)
	}
	if statusResp.Status != pb.NotificationStatus_SENT {
		t.Fatalf("Expected status SENT, got %v", statusResp.Status)
	}
}
