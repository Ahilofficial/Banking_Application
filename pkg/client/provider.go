package client

import (
	"banking-microservices/pkg/common/config"

	"github.com/google/wire"
)

// ProviderSet provides client dependencies
var ProviderSet = wire.NewSet(
	ProvideAuditClient,
	ProvideNotificationGRPCClient,
)

// ProvideAuditClient initializes an AuditClient using configuration
func ProvideAuditClient(cfg *config.Config) *AuditClient {
	return NewAuditClient(cfg.AuditServiceURL)
}

// ProvideNotificationGRPCClient initializes a NotificationGRPCClient using configuration
func ProvideNotificationGRPCClient(cfg *config.Config) (*NotificationGRPCClient, error) {
	return NewNotificationGRPCClient(cfg.NotificationGRPCAddr)
}
