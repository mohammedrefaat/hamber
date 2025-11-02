package dbmodels

import "time"

// ========== DASHBOARD STATISTICS ==========

// DashboardStats represents user dashboard statistics
type DashboardStats struct {
	UserID uint `json:"user_id"`

	// Orders
	TotalOrders     int64   `json:"total_orders"`
	PendingOrders   int64   `json:"pending_orders"`
	CompletedOrders int64   `json:"completed_orders"`
	TotalRevenue    float64 `json:"total_revenue"`
	MonthlyRevenue  float64 `json:"monthly_revenue"`

	// Products
	TotalProducts  int64 `json:"total_products"`
	ActiveProducts int64 `json:"active_products"`
	LowStockCount  int64 `json:"low_stock_count"`

	// Clients
	TotalClients    int64 `json:"total_clients"`
	NewClientsMonth int64 `json:"new_clients_month"`

	// Blog
	TotalBlogs     int64 `json:"total_blogs"`
	PublishedBlogs int64 `json:"published_blogs"`
	DraftBlogs     int64 `json:"draft_blogs"`

	// Messages
	UnreadMessages int64 `json:"unread_messages"`
	TotalMessages  int64 `json:"total_messages"`

	// Todos
	PendingTodos   int64 `json:"pending_todos"`
	CompletedTodos int64 `json:"completed_todos"`
	OverdueTodos   int64 `json:"overdue_todos"`

	// Calendar
	TodayEvents    int64 `json:"today_events"`
	UpcomingEvents int64 `json:"upcoming_events"`

	// Notifications
	UnreadNotifications int64 `json:"unread_notifications"`

	// Addons
	ActiveAddons int64 `json:"active_addons"`

	// Generated at
	GeneratedAt time.Time `json:"generated_at"`
}
