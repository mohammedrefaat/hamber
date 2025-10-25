package notification

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	"github.com/mohammedrefaat/hamber/stores"
	amqp "github.com/rabbitmq/amqp091-go"
)

// NotificationService handles all notification operations
type NotificationService struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	store     *stores.DbStore
	queueName string
}

// NotificationMessage represents a message in the queue
type NotificationMessage struct {
	UserID    uint                   `json:"user_id"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Type      string                 `json:"type"` // info, warning, error, success
	Link      string                 `json:"link"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// NewNotificationService creates a new notification service
func NewNotificationService(rabbitMQURL string, store *stores.DbStore) (*NotificationService, error) {
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	queueName := "notifications"

	// Declare queue
	_, err = channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %v", err)
	}

	service := &NotificationService{
		conn:      conn,
		channel:   channel,
		store:     store,
		queueName: queueName,
	}

	// Start consumer
	go service.startConsumer()

	log.Println("✓ Notification service initialized successfully")
	return service, nil
}

// PublishNotification publishes a notification to the queue
func (ns *NotificationService) PublishNotification(msg NotificationMessage) error {
	msg.CreatedAt = time.Now()

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	err = ns.channel.Publish(
		"",           // exchange
		ns.queueName, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	log.Printf("📤 Published notification for user %d: %s", msg.UserID, msg.Title)
	return nil
}

// PublishBulkNotifications publishes multiple notifications
func (ns *NotificationService) PublishBulkNotifications(userIDs []uint, title, message, notifType, link string) error {
	for _, userID := range userIDs {
		msg := NotificationMessage{
			UserID:  userID,
			Title:   title,
			Message: message,
			Type:    notifType,
			Link:    link,
		}

		if err := ns.PublishNotification(msg); err != nil {
			log.Printf("⚠️ Failed to publish notification for user %d: %v", userID, err)
			// Continue with other users
		}
	}
	return nil
}

// startConsumer starts consuming messages from the queue
func (ns *NotificationService) startConsumer() {
	msgs, err := ns.channel.Consume(
		ns.queueName, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		log.Printf("❌ Failed to register consumer: %v", err)
		return
	}

	log.Println("🔄 Notification consumer started")

	for msg := range msgs {
		if err := ns.processNotification(msg); err != nil {
			log.Printf("❌ Error processing notification: %v", err)
			msg.Nack(false, true) // Requeue message
		} else {
			msg.Ack(false) // Acknowledge message
		}
	}
}

// processNotification processes a notification message
func (ns *NotificationService) processNotification(msg amqp.Delivery) error {
	var notifMsg NotificationMessage
	if err := json.Unmarshal(msg.Body, &notifMsg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %v", err)
	}

	// Create notification in database
	notification := &dbmodels.Notification{
		UserID:  notifMsg.UserID,
		Title:   notifMsg.Title,
		Message: notifMsg.Message,
		Type:    notifMsg.Type,
		Link:    notifMsg.Link,
		IsRead:  false,
	}

	if err := ns.store.CreateNotification(notification); err != nil {
		return fmt.Errorf("failed to save notification: %v", err)
	}

	log.Printf("✅ Notification saved for user %d: %s", notifMsg.UserID, notifMsg.Title)

	// Here you can add additional handlers:
	// - Send push notification
	// - Send email
	// - Send SMS
	// - WebSocket notification

	return nil
}

// Close closes the notification service
func (ns *NotificationService) Close() {
	if ns.channel != nil {
		ns.channel.Close()
	}
	if ns.conn != nil {
		ns.conn.Close()
	}
	log.Println("🔒 Notification service closed")
}

// Helper functions for common notification scenarios

// NotifyNewOrder sends notification for new order
func (ns *NotificationService) NotifyNewOrder(userID uint, orderID uint, total float64) error {
	return ns.PublishNotification(NotificationMessage{
		UserID:  userID,
		Title:   "New Order Created",
		Message: fmt.Sprintf("Your order #%d has been created successfully. Total: %.2f EGP", orderID, total),
		Type:    "success",
		Link:    fmt.Sprintf("/orders/%d", orderID),
	})
}

// NotifyOrderStatusChange sends notification for order status change
func (ns *NotificationService) NotifyOrderStatusChange(userID uint, orderID uint, status string) error {
	return ns.PublishNotification(NotificationMessage{
		UserID:  userID,
		Title:   "Order Status Updated",
		Message: fmt.Sprintf("Your order #%d status has been updated to: %s", orderID, status),
		Type:    "info",
		Link:    fmt.Sprintf("/orders/%d", orderID),
	})
}

// NotifyPaymentSuccess sends notification for successful payment
func (ns *NotificationService) NotifyPaymentSuccess(userID uint, paymentID uint, amount float64) error {
	return ns.PublishNotification(NotificationMessage{
		UserID:  userID,
		Title:   "Payment Successful",
		Message: fmt.Sprintf("Your payment of %.2f EGP has been processed successfully", amount),
		Type:    "success",
		Link:    fmt.Sprintf("/payments/%d", paymentID),
	})
}

// NotifyPaymentFailed sends notification for failed payment
func (ns *NotificationService) NotifyPaymentFailed(userID uint, paymentID uint, reason string) error {
	return ns.PublishNotification(NotificationMessage{
		UserID:  userID,
		Title:   "Payment Failed",
		Message: fmt.Sprintf("Your payment failed: %s", reason),
		Type:    "error",
		Link:    fmt.Sprintf("/payments/%d", paymentID),
	})
}

// NotifyPackageChange sends notification for package change
func (ns *NotificationService) NotifyPackageChange(userID uint, oldPackage, newPackage string) error {
	return ns.PublishNotification(NotificationMessage{
		UserID:  userID,
		Title:   "Package Updated",
		Message: fmt.Sprintf("Your package has been updated from %s to %s", oldPackage, newPackage),
		Type:    "success",
		Link:    "/profile/subscription",
	})
}

// NotifyAddonSubscription sends notification for addon subscription
func (ns *NotificationService) NotifyAddonSubscription(userID uint, addonName string, expiryDate time.Time) error {
	return ns.PublishNotification(NotificationMessage{
		UserID:  userID,
		Title:   "Add-on Activated",
		Message: fmt.Sprintf("Your %s add-on has been activated. Valid until %s", addonName, expiryDate.Format("2006-01-02")),
		Type:    "success",
		Link:    "/profile/addons",
	})
}

// NotifyEventReminder sends notification for calendar event reminder
func (ns *NotificationService) NotifyEventReminder(userID uint, eventTitle string, startTime time.Time) error {
	return ns.PublishNotification(NotificationMessage{
		UserID:  userID,
		Title:   "Event Reminder",
		Message: fmt.Sprintf("Reminder: %s starts at %s", eventTitle, startTime.Format("15:04")),
		Type:    "info",
		Link:    "/calendar",
	})
}

// NotifyTodoDeadline sends notification for todo deadline
func (ns *NotificationService) NotifyTodoDeadline(userID uint, todoTitle string, dueDate time.Time) error {
	return ns.PublishNotification(NotificationMessage{
		UserID:  userID,
		Title:   "Task Deadline Approaching",
		Message: fmt.Sprintf("Task '%s' is due on %s", todoTitle, dueDate.Format("2006-01-02")),
		Type:    "warning",
		Link:    "/todos",
	})
}

// NotifySystemMaintenance sends notification for system maintenance
func (ns *NotificationService) NotifySystemMaintenance(userID uint, maintenanceTime time.Time, duration string) error {
	return ns.PublishNotification(NotificationMessage{
		UserID:  userID,
		Title:   "Scheduled Maintenance",
		Message: fmt.Sprintf("System maintenance scheduled for %s. Duration: %s", maintenanceTime.Format("2006-01-02 15:04"), duration),
		Type:    "warning",
		Link:    "",
	})
}

// NotifySecurityAlert sends notification for security issues
func (ns *NotificationService) NotifySecurityAlert(userID uint, alertMessage string) error {
	return ns.PublishNotification(NotificationMessage{
		UserID:  userID,
		Title:   "Security Alert",
		Message: alertMessage,
		Type:    "error",
		Link:    "/profile/security",
	})
}

// NotifyNewMessage sends notification for new message
func (ns *NotificationService) NotifyNewMessage(userID uint, senderName, messagePreview string) error {
	return ns.PublishNotification(NotificationMessage{
		UserID:  userID,
		Title:   "New Message",
		Message: fmt.Sprintf("%s sent you a message: %s", senderName, messagePreview),
		Type:    "info",
		Link:    "/messages",
	})
}

// NotifyWelcome sends welcome notification to new users
func (ns *NotificationService) NotifyWelcome(userID uint, userName string) error {
	return ns.PublishNotification(NotificationMessage{
		UserID:  userID,
		Title:   "Welcome to Hamber!",
		Message: fmt.Sprintf("Hello %s! Welcome to Hamber platform. Let's get you started!", userName),
		Type:    "info",
		Link:    "/getting-started",
	})
}
