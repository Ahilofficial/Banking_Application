package notification

import (
	"fmt"
)

type NotificationType int32

const (
	NotificationType_NOTIFICATION_TYPE_UNSPECIFIED NotificationType = 0
	NotificationType_EMAIL                         NotificationType = 1
	NotificationType_SMS                           NotificationType = 2
	NotificationType_PUSH                          NotificationType = 3
	NotificationType_IN_APP                        NotificationType = 4
)

var NotificationType_name = map[int32]string{
	0: "NOTIFICATION_TYPE_UNSPECIFIED",
	1: "EMAIL",
	2: "SMS",
	3: "PUSH",
	4: "IN_APP",
}

var NotificationType_value = map[string]int32{
	"NOTIFICATION_TYPE_UNSPECIFIED": 0,
	"EMAIL":                         1,
	"SMS":                           2,
	"PUSH":                          3,
	"IN_APP":                        4,
}

func (x NotificationType) String() string {
	if s, ok := NotificationType_name[int32(x)]; ok {
		return s
	}
	return fmt.Sprintf("NotificationType(%d)", x)
}

type NotificationStatus int32

const (
	NotificationStatus_NOTIFICATION_STATUS_UNSPECIFIED NotificationStatus = 0
	NotificationStatus_PENDING                         NotificationStatus = 1
	NotificationStatus_SENT                            NotificationStatus = 2
	NotificationStatus_DELIVERED                       NotificationStatus = 3
	NotificationStatus_FAILED                          NotificationStatus = 4
)

var NotificationStatus_name = map[int32]string{
	0: "NOTIFICATION_STATUS_UNSPECIFIED",
	1: "PENDING",
	2: "SENT",
	3: "DELIVERED",
	4: "FAILED",
}

var NotificationStatus_value = map[string]int32{
	"NOTIFICATION_STATUS_UNSPECIFIED": 0,
	"PENDING":                         1,
	"SENT":                            2,
	"DELIVERED":                       3,
	"FAILED":                          4,
}

func (x NotificationStatus) String() string {
	if s, ok := NotificationStatus_name[int32(x)]; ok {
		return s
	}
	return fmt.Sprintf("NotificationStatus(%d)", x)
}

// SendEmailRequest represents email payload
type SendEmailRequest struct {
	UserId   uint32            `json:"user_id,omitempty"`
	To       string            `json:"to,omitempty"`
	Subject  string            `json:"subject,omitempty"`
	Body     string            `json:"body,omitempty"`
	HtmlBody string            `json:"html_body,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func (m *SendEmailRequest) Reset()         { *m = SendEmailRequest{} }
func (m *SendEmailRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*SendEmailRequest) ProtoMessage()    {}

// SendEmailResponse represents email delivery response
type SendEmailResponse struct {
	NotificationId string             `json:"notification_id,omitempty"`
	Status         NotificationStatus `json:"status,omitempty"`
	Message        string             `json:"message,omitempty"`
	SentAt         string             `json:"sent_at,omitempty"`
}

func (m *SendEmailResponse) Reset()         { *m = SendEmailResponse{} }
func (m *SendEmailResponse) String() string { return fmt.Sprintf("%+v", *m) }
func (*SendEmailResponse) ProtoMessage()    {}

// SendSMSRequest represents SMS payload
type SendSMSRequest struct {
	UserId      uint32            `json:"user_id,omitempty"`
	PhoneNumber string            `json:"phone_number,omitempty"`
	Message     string            `json:"message,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

func (m *SendSMSRequest) Reset()         { *m = SendSMSRequest{} }
func (m *SendSMSRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*SendSMSRequest) ProtoMessage()    {}

// SendSMSResponse represents SMS response
type SendSMSResponse struct {
	NotificationId string             `json:"notification_id,omitempty"`
	Status         NotificationStatus `json:"status,omitempty"`
	Message        string             `json:"message,omitempty"`
	SentAt         string             `json:"sent_at,omitempty"`
}

func (m *SendSMSResponse) Reset()         { *m = SendSMSResponse{} }
func (m *SendSMSResponse) String() string { return fmt.Sprintf("%+v", *m) }
func (*SendSMSResponse) ProtoMessage()    {}

// SendNotificationRequest represents push/in-app notification payload
type SendNotificationRequest struct {
	UserId    uint32            `json:"user_id,omitempty"`
	Type      NotificationType  `json:"type,omitempty"`
	Recipient string            `json:"recipient,omitempty"`
	Title     string            `json:"title,omitempty"`
	Content   string            `json:"content,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

func (m *SendNotificationRequest) Reset()         { *m = SendNotificationRequest{} }
func (m *SendNotificationRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*SendNotificationRequest) ProtoMessage()    {}

// SendNotificationResponse represents notification response
type SendNotificationResponse struct {
	NotificationId string             `json:"notification_id,omitempty"`
	Status         NotificationStatus `json:"status,omitempty"`
	Message        string             `json:"message,omitempty"`
	SentAt         string             `json:"sent_at,omitempty"`
}

func (m *SendNotificationResponse) Reset()         { *m = SendNotificationResponse{} }
func (m *SendNotificationResponse) String() string { return fmt.Sprintf("%+v", *m) }
func (*SendNotificationResponse) ProtoMessage()    {}

// GetNotificationsRequest query filters
type GetNotificationsRequest struct {
	UserId uint32           `json:"user_id,omitempty"`
	Type   NotificationType `json:"type,omitempty"`
	Limit  int32            `json:"limit,omitempty"`
	Offset int32            `json:"offset,omitempty"`
}

func (m *GetNotificationsRequest) Reset()         { *m = GetNotificationsRequest{} }
func (m *GetNotificationsRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*GetNotificationsRequest) ProtoMessage()    {}

// NotificationItem single notification record
type NotificationItem struct {
	Id        string             `json:"id,omitempty"`
	UserId    uint32             `json:"user_id,omitempty"`
	Type      NotificationType   `json:"type,omitempty"`
	Recipient string             `json:"recipient,omitempty"`
	Title     string             `json:"title,omitempty"`
	Content   string             `json:"content,omitempty"`
	Status    NotificationStatus `json:"status,omitempty"`
	CreatedAt string             `json:"created_at,omitempty"`
}

func (m *NotificationItem) Reset()         { *m = NotificationItem{} }
func (m *NotificationItem) String() string { return fmt.Sprintf("%+v", *m) }
func (*NotificationItem) ProtoMessage()    {}

// GetNotificationsResponse list of notifications
type GetNotificationsResponse struct {
	Notifications []*NotificationItem `json:"notifications,omitempty"`
	Total         int32               `json:"total,omitempty"`
}

func (m *GetNotificationsResponse) Reset()         { *m = GetNotificationsResponse{} }
func (m *GetNotificationsResponse) String() string { return fmt.Sprintf("%+v", *m) }
func (*GetNotificationsResponse) ProtoMessage()    {}

// GetNotificationStatusRequest request status by ID
type GetNotificationStatusRequest struct {
	NotificationId string `json:"notification_id,omitempty"`
}

func (m *GetNotificationStatusRequest) Reset()         { *m = GetNotificationStatusRequest{} }
func (m *GetNotificationStatusRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*GetNotificationStatusRequest) ProtoMessage()    {}

// NotificationStatusResponse status inquiry response
type NotificationStatusResponse struct {
	NotificationId string             `json:"notification_id,omitempty"`
	Status         NotificationStatus `json:"status,omitempty"`
	ErrorMessage   string             `json:"error_message,omitempty"`
	UpdatedAt      string             `json:"updated_at,omitempty"`
}

func (m *NotificationStatusResponse) Reset()         { *m = NotificationStatusResponse{} }
func (m *NotificationStatusResponse) String() string { return fmt.Sprintf("%+v", *m) }
func (*NotificationStatusResponse) ProtoMessage()    {}
