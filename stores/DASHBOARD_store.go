package stores

import (
	"net/http"
	"time"

	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
)

// ========== USER DASHBOARD STATISTICS ==========

func (store *DbStore) GetDashboardStats(userID uint) (*dbmodels.DashboardStats, error) {
	stats := &dbmodels.DashboardStats{
		UserID:      userID,
		GeneratedAt: time.Now(),
	}

	// Orders
	store.db.Model(&dbmodels.Order{}).Where("user_id = ?", userID).Count(&stats.TotalOrders)
	store.db.Model(&dbmodels.Order{}).Where("user_id = ? AND status = ?", userID, dbmodels.OrderStatus_PENDING).Count(&stats.PendingOrders)
	store.db.Model(&dbmodels.Order{}).Where("user_id = ? AND status = ?", userID, dbmodels.OrderStatus_DELIVERED).Count(&stats.CompletedOrders)

	// Total Revenue
	store.db.Model(&dbmodels.Order{}).Where("user_id = ?", userID).Select("COALESCE(SUM(total), 0)").Scan(&stats.TotalRevenue)

	// Monthly Revenue
	startOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	store.db.Model(&dbmodels.Order{}).Where("user_id = ? AND created_at >= ?", userID, startOfMonth).Select("COALESCE(SUM(total), 0)").Scan(&stats.MonthlyRevenue)

	// Products
	store.db.Model(&dbmodels.Product{}).Where("user_id = ?", userID).Count(&stats.TotalProducts)
	store.db.Model(&dbmodels.Product{}).Where("user_id = ? AND is_active = ?", userID, true).Count(&stats.ActiveProducts)
	store.db.Model(&dbmodels.Product{}).Where("user_id = ? AND quantity < ?", userID, 10).Count(&stats.LowStockCount)

	// Clients
	store.db.Model(&dbmodels.Client{}).Where("user_id = ?", userID).Count(&stats.TotalClients)
	store.db.Model(&dbmodels.Client{}).Where("user_id = ? AND created_at >= ?", userID, startOfMonth).Count(&stats.NewClientsMonth)

	// Blog
	store.db.Model(&dbmodels.Blog{}).Where("author_id = ?", userID).Count(&stats.TotalBlogs)
	store.db.Model(&dbmodels.Blog{}).Where("author_id = ? AND is_published = ?", userID, true).Count(&stats.PublishedBlogs)
	store.db.Model(&dbmodels.Blog{}).Where("author_id = ? AND is_published = ?", userID, false).Count(&stats.DraftBlogs)

	// Messages
	store.db.Model(&dbmodels.Message{}).Where("receiver_id = ? AND is_read = ? AND deleted_by_receiver = ?", userID, false, false).Count(&stats.UnreadMessages)
	store.db.Model(&dbmodels.Message{}).Where("receiver_id = ? AND deleted_by_receiver = ?", userID, false).Count(&stats.TotalMessages)

	// Todos
	store.db.Model(&dbmodels.Todo{}).Where("user_id = ? AND is_completed = ?", userID, false).Count(&stats.PendingTodos)
	store.db.Model(&dbmodels.Todo{}).Where("user_id = ? AND is_completed = ?", userID, true).Count(&stats.CompletedTodos)

	// Overdue todos
	store.db.Model(&dbmodels.Todo{}).Where("user_id = ? AND is_completed = ? AND due_date < ?", userID, false, time.Now()).Count(&stats.OverdueTodos)

	// Calendar
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.AddDate(0, 0, 1)
	store.db.Model(&dbmodels.CalendarEvent{}).Where("user_id = ? AND start_time >= ? AND start_time < ?", userID, today, tomorrow).Count(&stats.TodayEvents)

	nextWeek := time.Now().AddDate(0, 0, 7)
	store.db.Model(&dbmodels.CalendarEvent{}).Where("user_id = ? AND start_time >= ? AND start_time <= ?", userID, time.Now(), nextWeek).Count(&stats.UpcomingEvents)

	// Notifications
	store.db.Model(&dbmodels.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&stats.UnreadNotifications)

	// Addons
	store.db.Model(&dbmodels.UserAddonSubscription{}).Where("user_id = ? AND status = ?", userID, dbmodels.AddonSubscriptionStatus_ACTIVE).Count(&stats.ActiveAddons)

	return stats, nil
}

func (store *DbStore) GetRevenueChartData(userID uint) ([]map[string]interface{}, error) {
	var chartData []map[string]interface{}

	// Get revenue for last 12 months
	query := `
		SELECT 
			TO_CHAR(created_at, 'Mon YYYY') as month,
			COALESCE(SUM(total), 0) as revenue,
			COUNT(*) as order_count
		FROM orders
		WHERE user_id = ? AND created_at >= ?
		GROUP BY TO_CHAR(created_at, 'Mon YYYY'), DATE_TRUNC('month', created_at)
		ORDER BY DATE_TRUNC('month', created_at) ASC
		LIMIT 12
	`

	startDate := time.Now().AddDate(-1, 0, 0)
	if err := store.db.Raw(query, userID, startDate).Scan(&chartData).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to fetch revenue chart data",
			Code:    http.StatusInternalServerError,
		}
	}

	return chartData, nil
}

func (store *DbStore) GetOrdersChartData(userID uint) (map[string]int64, error) {
	data := make(map[string]int64)

	statuses := []dbmodels.OrderStatus{
		dbmodels.OrderStatus_PENDING,
		dbmodels.OrderStatus_SHIPPED,
		dbmodels.OrderStatus_DELIVERED,
		dbmodels.OrderStatus_CANCELED,
	}

	for _, status := range statuses {
		var count int64
		store.db.Model(&dbmodels.Order{}).Where("user_id = ? AND status = ?", userID, status).Count(&count)
		data[status.String()] = count
	}

	return data, nil
}

func (store *DbStore) GetTopSellingProducts(userID uint, limit int) ([]map[string]interface{}, error) {
	var products []map[string]interface{}

	query := `
		SELECT 
			p.id,
			p.name,
			p.price,
			COALESCE(SUM(oi.quantity), 0) as total_sold,
			COALESCE(SUM(oi.quantity * oi.price), 0) as total_revenue
		FROM products p
		LEFT JOIN order_items oi ON p.id = oi.product_id
		WHERE p.user_id = ?
		GROUP BY p.id, p.name, p.price
		ORDER BY total_sold DESC
		LIMIT ?
	`

	if err := store.db.Raw(query, userID, limit).Scan(&products).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to fetch top products",
			Code:    http.StatusInternalServerError,
		}
	}

	return products, nil
}

func (store *DbStore) GetRecentOrders(userID uint, limit int) ([]dbmodels.Order, error) {
	var orders []dbmodels.Order
	if err := store.db.Preload("Client").Preload("Items").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&orders).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to fetch recent orders",
			Code:    http.StatusInternalServerError,
		}
	}
	return orders, nil
}

func (store *DbStore) GetRecentActivities(userID uint, limit int) ([]map[string]interface{}, error) {
	var activities []map[string]interface{}

	// Combine recent activities from different sources
	query := `
		SELECT 'order' as type, id, created_at, 'New order created' as description FROM orders WHERE user_id = ?
		UNION ALL
		SELECT 'blog' as type, id, created_at, 'Blog post ' || CASE WHEN is_published THEN 'published' ELSE 'drafted' END as description FROM blogs WHERE author_id = ?
		UNION ALL
		SELECT 'product' as type, id, created_at, 'Product added' as description FROM products WHERE user_id = ?
		UNION ALL
		SELECT 'todo' as type, id, created_at, 'Todo ' || CASE WHEN is_completed THEN 'completed' ELSE 'created' END as description FROM todos WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	if err := store.db.Raw(query, userID, userID, userID, userID, limit).Scan(&activities).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to fetch activities",
			Code:    http.StatusInternalServerError,
		}
	}

	return activities, nil
}

func (store *DbStore) GetProductStatistics(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total products
	var total int64
	store.db.Model(&dbmodels.Product{}).Where("user_id = ?", userID).Count(&total)
	stats["total"] = total

	// By category
	var byCategory []map[string]interface{}
	store.db.Model(&dbmodels.Product{}).
		Select("category, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("category").
		Scan(&byCategory)
	stats["by_category"] = byCategory

	// Low stock
	var lowStock int64
	store.db.Model(&dbmodels.Product{}).Where("user_id = ? AND quantity < ?", userID, 10).Count(&lowStock)
	stats["low_stock"] = lowStock

	// Out of stock
	var outOfStock int64
	store.db.Model(&dbmodels.Product{}).Where("user_id = ? AND quantity = ?", userID, 0).Count(&outOfStock)
	stats["out_of_stock"] = outOfStock

	// Total inventory value
	var totalValue float64
	store.db.Model(&dbmodels.Product{}).Where("user_id = ?", userID).Select("COALESCE(SUM(price * quantity), 0)").Scan(&totalValue)
	stats["total_value"] = totalValue

	return stats, nil
}

func (store *DbStore) GetClientStatistics(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total clients
	var total int64
	store.db.Model(&dbmodels.Client{}).Where("user_id = ?", userID).Count(&total)
	stats["total"] = total

	// New this month
	startOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	var newThisMonth int64
	store.db.Model(&dbmodels.Client{}).Where("user_id = ? AND created_at >= ?", userID, startOfMonth).Count(&newThisMonth)
	stats["new_this_month"] = newThisMonth

	// Top clients by orders
	var topClients []map[string]interface{}
	query := `
		SELECT 
			c.id,
			c.name,
			c.email,
			COUNT(o.id) as order_count,
			COALESCE(SUM(o.total), 0) as total_spent
		FROM clients c
		LEFT JOIN orders o ON c.id = o.client_id
		WHERE c.user_id = ?
		GROUP BY c.id, c.name, c.email
		ORDER BY order_count DESC
		LIMIT 10
	`
	store.db.Raw(query, userID).Scan(&topClients)
	stats["top_clients"] = topClients

	return stats, nil
}

// ========== ADMIN DASHBOARD ==========

func (store *DbStore) GetAdminDashboardStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Users
	var totalUsers, activeUsers, newUsersToday int64
	store.db.Model(&dbmodels.User{}).Count(&totalUsers)
	store.db.Model(&dbmodels.User{}).Where("is_active = ?", true).Count(&activeUsers)
	today := time.Now().Truncate(24 * time.Hour)
	store.db.Model(&dbmodels.User{}).Where("created_at >= ?", today).Count(&newUsersToday)

	stats["total_users"] = totalUsers
	stats["active_users"] = activeUsers
	stats["new_users_today"] = newUsersToday

	// Orders
	var totalOrders int64
	var totalRevenue float64
	store.db.Model(&dbmodels.Order{}).Count(&totalOrders)
	store.db.Model(&dbmodels.Order{}).Select("COALESCE(SUM(total), 0)").Scan(&totalRevenue)

	stats["total_orders"] = totalOrders
	stats["total_revenue"] = totalRevenue

	// Products
	var totalProducts int64
	store.db.Model(&dbmodels.Product{}).Count(&totalProducts)
	stats["total_products"] = totalProducts

	// Blogs
	var totalBlogs, publishedBlogs int64
	store.db.Model(&dbmodels.Blog{}).Count(&totalBlogs)
	store.db.Model(&dbmodels.Blog{}).Where("is_published = ?", true).Count(&publishedBlogs)

	stats["total_blogs"] = totalBlogs
	stats["published_blogs"] = publishedBlogs

	// Addons
	var activeSubscriptions int64
	store.db.Model(&dbmodels.UserAddonSubscription{}).Where("status = ?", dbmodels.AddonSubscriptionStatus_ACTIVE).Count(&activeSubscriptions)
	stats["active_subscriptions"] = activeSubscriptions

	return stats, nil
}

func (store *DbStore) GetPlatformAnalytics() (map[string]interface{}, error) {
	analytics := make(map[string]interface{})

	// User distribution by package
	var byPackage []map[string]interface{}
	store.db.Model(&dbmodels.User{}).
		Select("package_id, COUNT(*) as count").
		Group("package_id").
		Scan(&byPackage)
	analytics["users_by_package"] = byPackage

	// Revenue by month (last 12 months)
	var revenueByMonth []map[string]interface{}
	query := `
		SELECT 
			TO_CHAR(created_at, 'Mon YYYY') as month,
			COALESCE(SUM(total), 0) as revenue
		FROM orders
		WHERE created_at >= ?
		GROUP BY TO_CHAR(created_at, 'Mon YYYY'), DATE_TRUNC('month', created_at)
		ORDER BY DATE_TRUNC('month', created_at) ASC
	`
	startDate := time.Now().AddDate(-1, 0, 0)
	store.db.Raw(query, startDate).Scan(&revenueByMonth)
	analytics["revenue_by_month"] = revenueByMonth

	return analytics, nil
}

func (store *DbStore) GetUserGrowthChart() ([]map[string]interface{}, error) {
	var chartData []map[string]interface{}

	query := `
		SELECT 
			TO_CHAR(created_at, 'Mon YYYY') as month,
			COUNT(*) as user_count
		FROM users
		WHERE created_at >= ?
		GROUP BY TO_CHAR(created_at, 'Mon YYYY'), DATE_TRUNC('month', created_at)
		ORDER BY DATE_TRUNC('month', created_at) ASC
	`

	startDate := time.Now().AddDate(-1, 0, 0)
	if err := store.db.Raw(query, startDate).Scan(&chartData).Error; err != nil {
		return nil, &CustomError{
			Message: "Failed to fetch user growth data",
			Code:    http.StatusInternalServerError,
		}
	}

	return chartData, nil
}

func (store *DbStore) GetRevenueBreakdown() (map[string]interface{}, error) {
	breakdown := make(map[string]interface{})

	// Revenue by payment status
	var byPaymentStatus []map[string]interface{}
	store.db.Model(&dbmodels.Order{}).
		Select("payment_status, COALESCE(SUM(total), 0) as revenue, COUNT(*) as count").
		Group("payment_status").
		Scan(&byPaymentStatus)
	breakdown["by_payment_status"] = byPaymentStatus

	// Revenue by order status
	var byOrderStatus []map[string]interface{}
	store.db.Model(&dbmodels.Order{}).
		Select("status, COALESCE(SUM(total), 0) as revenue, COUNT(*) as count").
		Group("status").
		Scan(&byOrderStatus)
	breakdown["by_order_status"] = byOrderStatus

	return breakdown, nil
}
