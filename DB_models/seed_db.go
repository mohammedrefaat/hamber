package dbmodels

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// SeedDatabase fills the database with dummy data for frontend testing
func SeedDatabase(db *gorm.DB) error {
	fmt.Println("🌱 Starting database seeding...")
	// Run migrations first
	if err := Migrator(db); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Seed in order due to foreign key constraints
	/*if err := seedRolesAndPermissions(db); err != nil {
		return err
	}
	if err := seedPackages(db); err != nil {
		return err
	}
	if err := seedUsers(db); err != nil {
		return err
	}
	if err := seedClients(db); err != nil {
		return err
	}
	if err := seedProducts(db); err != nil {
		return err
	}
	if err := seedOrders(db); err != nil {
		return err
	}
	if err := seedAddons(db); err != nil {
		return err
	}
	if err := seedPayments(db); err != nil {
		return err
	}
	if err := seedAddonSubscriptions(db); err != nil {
		return err
	}

	if err := seedTodos(db); err != nil {
		return err
	}
	if err := seedCalendarEvents(db); err != nil {
		return err
	}
	if err := seedNotifications(db); err != nil {
		return err
	}
	if err := seedBlogs(db); err != nil {
		return err
	}
	if err := seedNewsletters(db); err != nil {
		return err
	}
	if err := seedContacts(db); err != nil {
		return err
	}*/
	if err := seedReceipts(db); err != nil {
		return err
	}
	fmt.Println("✅ Database seeding completed successfully!")
	return nil
}

func seedRolesAndPermissions(db *gorm.DB) error {
	fmt.Println("📝 Seeding roles and permissions...")

	permissionNames := []string{
		"CREATE_ORDER",
		"VIEW_ORDER",
		"UPDATE_ORDER",
		"DELETE_ORDER",
		"MANAGE_USERS",
		"MANAGE_PRODUCTS",
		"VIEW_REPORTS",
		"MANAGE_PACKAGES",
	}

	var permissions []Permission
	for d, name := range permissionNames {
		var perm Permission
		result := db.Where("name = ?", name).First(&perm)
		if result.Error != nil {
			// Create if doesn't exist
			perm = Permission{Name: name, ID: uint(d + 1), CreatedAt: time.Now(), UpdatedAt: time.Now()}
			if err := db.Save(&perm).Error; err != nil {
				return fmt.Errorf("failed to seed permission: %w", err)
			}
		}
		permissions = append(permissions, perm)
	}

	roleNames := []string{"Admin", "Manager", "User", "Client"}

	for _, name := range roleNames {
		var role Role
		result := db.Where("name = ?", name).First(&role)
		if result.Error != nil {
			// Create if doesn't exist
			role = Role{Name: name}
			if err := db.Save(&role).Error; err != nil {
				return fmt.Errorf("failed to seed role: %w", err)
			}
		}
	}

	// Assign permissions to roles
	var adminRole Role
	db.Where("name = ?", "Admin").First(&adminRole)
	if adminRole.ID > 0 {
		db.Model(&adminRole).Association("Permissions").Replace(permissions)
	}

	var managerRole Role
	db.Where("name = ?", "Manager").First(&managerRole)
	if managerRole.ID > 0 {
		db.Model(&managerRole).Association("Permissions").Replace(permissions[:6])
	}

	fmt.Println("✓ Roles and permissions seeded")
	return nil
}

func seedPackages(db *gorm.DB) error {
	fmt.Println("📦 Seeding packages...")

	benefits1, _ := json.Marshal([]string{"5 Users", "50 Products", "Email Support", "Basic Analytics"})
	benefits2, _ := json.Marshal([]string{"15 Users", "200 Products", "Priority Support", "Advanced Analytics", "API Access"})
	benefits3, _ := json.Marshal([]string{"Unlimited Users", "Unlimited Products", "24/7 Premium Support", "Custom Integrations", "Dedicated Account Manager"})

	packages := []Package{
		{
			Name:           "Starter",
			Price:          299.99,
			Duration:       30,
			Benefits:       string(benefits1),
			Description:    "Perfect for small businesses just getting started",
			IsActive:       true,
			PricePerClient: false,
		},
		{
			Name:           "Professional",
			Price:          899.99,
			Duration:       30,
			Benefits:       string(benefits2),
			Description:    "Ideal for growing businesses with advanced needs",
			IsActive:       true,
			PricePerClient: false,
		},
		{
			Name:           "Enterprise",
			Price:          2499.99,
			Duration:       30,
			Benefits:       string(benefits3),
			Description:    "Full-featured solution for large organizations",
			IsActive:       true,
			PricePerClient: true,
		},
	}

	for i := range packages {
		var existing Package
		result := db.Where("name = ?", packages[i].Name).First(&existing)
		if result.Error != nil {
			// Create if doesn't exist
			if err := db.Save(&packages[i]).Error; err != nil {
				return fmt.Errorf("failed to seed package: %w", err)
			}
		}
	}

	fmt.Println("✓ Packages seeded")
	return nil
}

func seedUsers(db *gorm.DB) error {
	fmt.Println("👥 Seeding users...")

	var adminRole Role
	db.Where("name = ?", "Admin").First(&adminRole)

	var userRole Role
	db.Where("name = ?", "User").First(&userRole)

	var pkg1, pkg2, pkg3 Package
	db.First(&pkg1, 1)
	db.First(&pkg2, 2)
	db.First(&pkg3, 3)

	users := []User{
		{
			Name:               "John Smith",
			Email:              "john.smith@example.com",
			Phone:              "+201001234567",
			Subdomain:          "johnsmith",
			RoleID:             adminRole.ID,
			PackageID:          pkg3.ID,
			IS_ACTIVE:          true,
			IS_EMAIL_VERIFIED:  true,
			IS_MOBILE_VERIFIED: true,
			Avatar:             "https://i.pravatar.cc/150?img=12",
			Bio:                "Senior business consultant with 10+ years experience",
			Location:           "Cairo, Egypt",
		},
		{
			Name:               "Sarah Johnson",
			Email:              "sarah.j@example.com",
			Phone:              "+201009876543",
			Subdomain:          "sarahj",
			RoleID:             userRole.ID,
			PackageID:          pkg2.ID,
			IS_ACTIVE:          true,
			IS_EMAIL_VERIFIED:  true,
			IS_MOBILE_VERIFIED: false,
			Avatar:             "https://i.pravatar.cc/150?img=5",
			Bio:                "E-commerce entrepreneur",
			Location:           "Alexandria, Egypt",
		},
		{
			Name:               "Ahmed Hassan",
			Email:              "ahmed.hassan@example.com",
			Phone:              "+201112345678",
			Subdomain:          "ahmedh",
			RoleID:             userRole.ID,
			PackageID:          pkg1.ID,
			IS_ACTIVE:          true,
			IS_EMAIL_VERIFIED:  true,
			IS_MOBILE_VERIFIED: true,
			Avatar:             "https://i.pravatar.cc/150?img=33",
			Bio:                "Digital marketing specialist",
			Location:           "Giza, Egypt",
		},
		{
			Name:               "Emily Davis",
			Email:              "emily.davis@example.com",
			Phone:              "+201098765432",
			Subdomain:          "emilyd",
			RoleID:             userRole.ID,
			PackageID:          pkg2.ID,
			IS_ACTIVE:          true,
			IS_EMAIL_VERIFIED:  false,
			IS_MOBILE_VERIFIED: false,
			Avatar:             "https://i.pravatar.cc/150?img=20",
			Location:           "Hurghada, Egypt",
		},
		{
			Name:               "Mohamed Ali",
			Email:              "mohamed.ali@example.com",
			Phone:              "+201234567890",
			Subdomain:          "mohameda",
			RoleID:             userRole.ID,
			PackageID:          pkg1.ID,
			IS_ACTIVE:          false,
			IS_EMAIL_VERIFIED:  true,
			IS_MOBILE_VERIFIED: true,
			Avatar:             "https://i.pravatar.cc/150?img=15",
			Bio:                "Software developer and tech enthusiast",
			Location:           "Mansoura, Egypt",
		},
	}

	for i := range users {
		var existing User
		result := db.Where("email = ?", users[i].Email).First(&existing)
		if result.Error != nil {
			// Create if doesn't exist
			if err := users[i].HashPassword("Password123!"); err != nil {
				return fmt.Errorf("failed to hash password: %w", err)
			}
			if err := db.Save(&users[i]).Error; err != nil {
				return fmt.Errorf("failed to seed user: %w", err)
			}
		}
	}

	fmt.Println("✓ Users seeded")
	return nil
}

func seedClients(db *gorm.DB) error {
	fmt.Println("👔 Seeding clients...")

	clients := []Client{
		{Name: "Acme Corporation", Email: "contact@acme.com", UserID: 2},
		{Name: "TechStart Inc", Email: "info@techstart.com", UserID: 2},
		{Name: "Global Solutions", Email: "hello@globalsol.com", UserID: 3},
		{Name: "Retail Pro LLC", Email: "support@retailpro.com", UserID: 3},
		{Name: "Digital Wave", Email: "contact@digitalwave.com", UserID: 3},
		{Name: "Innovation Hub", Email: "info@innovhub.com", UserID: 3},
		{Name: "Smart Systems", Email: "hello@smartsys.com", UserID: 2},
		{Name: "Future Tech", Email: "contact@futuretech.com", UserID: 3},
	}

	for i := range clients {
		var existing Client
		result := db.Where("email = ?", clients[i].Email).First(&existing)
		if result.Error != nil {
			// Create if doesn't exist
			if err := db.Save(&clients[i]).Error; err != nil {
				return fmt.Errorf("failed to seed client: %w", err)
			}
		}
	}

	fmt.Println("✓ Clients seeded")
	return nil
}

func seedProducts(db *gorm.DB) error {
	fmt.Println("📦 Seeding products...")

	images1, _ := json.Marshal([]string{"https://picsum.photos/400/400?random=1", "https://picsum.photos/400/400?random=2"})
	images2, _ := json.Marshal([]string{"https://picsum.photos/400/400?random=3"})
	images3, _ := json.Marshal([]string{"https://picsum.photos/400/400?random=4", "https://picsum.photos/400/400?random=5", "https://picsum.photos/400/400?random=6"})

	tags1, _ := json.Marshal([]string{"electronics", "smartphone", "bestseller"})
	tags2, _ := json.Marshal([]string{"laptop", "business", "premium"})
	tags3, _ := json.Marshal([]string{"accessories", "wireless"})

	products := []Product{
		{
			Name:          "Premium Smartphone X1",
			Description:   "Latest flagship smartphone with advanced features",
			Price:         15999.99,
			DiscountPrice: 13999.99,
			Quantity:      50,
			SKU:           "PHONE-X1-001",
			Category:      "Electronics",
			Brand:         "TechBrand",
			Images:        string(images1),
			IsActive:      true,
			Weight:        0.180,
			Tags:          string(tags1),
			UserID:        2,
			Favorite:      true,
		},
		{
			Name:          "Professional Laptop Pro 15",
			Description:   "High-performance laptop for professionals",
			Price:         32999.99,
			DiscountPrice: 29999.99,
			Quantity:      25,
			SKU:           "LAPTOP-PRO15-001",
			Category:      "Computers",
			Brand:         "TechBrand",
			Images:        string(images2),
			IsActive:      true,
			Weight:        2.1,
			Tags:          string(tags2),
			UserID:        2,
		},
		{
			Name:        "Wireless Earbuds Elite",
			Description: "Premium wireless earbuds with noise cancellation",
			Price:       2499.99,
			Quantity:    100,
			SKU:         "EARBUDS-ELITE-001",
			Category:    "Audio",
			Brand:       "AudioMax",
			Images:      string(images3),
			IsActive:    true,
			Weight:      0.05,
			Tags:        string(tags3),
			UserID:      2,
			Favorite:    false,
		},
		{
			Name:          "Smart Watch Series 5",
			Description:   "Feature-rich smartwatch with health tracking",
			Price:         4999.99,
			DiscountPrice: 4499.99,
			Quantity:      75,
			SKU:           "WATCH-S5-001",
			Category:      "Wearables",
			Brand:         "TechBrand",
			Images:        string(images1),
			IsActive:      true,
			Weight:        0.035,
			Tags:          string(tags1),
			UserID:        2,
		},
		{
			Name:        "Portable Charger 20000mAh",
			Description: "High-capacity power bank for mobile devices",
			Price:       799.99,
			Quantity:    200,
			SKU:         "CHARGER-20K-001",
			Category:    "Accessories",
			Brand:       "PowerPlus",
			Images:      string(images2),
			IsActive:    true,
			Weight:      0.4,
			Tags:        string(tags3),
			UserID:      3,
		},
	}

	for i := range products {
		var existing Product
		result := db.Where("sku = ?", products[i].SKU).First(&existing)
		if result.Error != nil {
			// Create if doesn't exist
			if err := db.Save(&products[i]).Error; err != nil {
				return fmt.Errorf("failed to seed product: %w", err)
			}
		}
	}

	fmt.Println("✓ Products seeded")
	return nil
}

func seedOrders(db *gorm.DB) error {
	fmt.Println("🛒 Seeding orders...")

	now := time.Now()
	paymentDate1 := now.Add(-24 * time.Hour)
	paymentDate2 := now.Add(-48 * time.Hour)

	orders := []Order{
		{
			ClientID:          1,
			UserID:            2,
			Total:             18999.98,
			Status:            OrderStatus_DELIVERED,
			Address:           "123 Main Street, Cairo, Egypt",
			Phone:             "+201001234567",
			Notes:             "Please call before delivery",
			PaymentStatus:     "PAID",
			PaymentAmount:     18999.98,
			PaymentMethodId:   1,
			PaymentMethodDesc: "Credit Card",
			PaymentDate:       &paymentDate1,
			PaymentRef:        "PAY-2024-001",
		},
		{
			ClientID:          2,
			UserID:            2,
			Total:             32999.99,
			Status:            OrderStatus_SHIPPED,
			Address:           "456 Business Ave, Alexandria, Egypt",
			Phone:             "+201009876543",
			PaymentStatus:     "PAID",
			PaymentAmount:     32999.99,
			PaymentMethodId:   2,
			PaymentMethodDesc: "Fawry",
			PaymentDate:       &paymentDate2,
			PaymentRef:        "PAY-2024-002",
		},
		{
			ClientID:      3,
			UserID:        2,
			Total:         2499.99,
			Status:        OrderStatus_PENDING,
			Address:       "789 Tech Street, Giza, Egypt",
			Phone:         "+201112345678",
			PaymentStatus: "PENDING",
		},
		{
			ClientID:          4,
			UserID:            2,
			Total:             5299.98,
			Status:            OrderStatus_DELIVERED,
			Address:           "321 Commerce Blvd, Mansoura, Egypt",
			Phone:             "+201098765432",
			Notes:             "Leave at reception",
			PaymentStatus:     "PAID",
			PaymentAmount:     5299.98,
			PaymentMethodId:   1,
			PaymentMethodDesc: "Paymob",
			PaymentDate:       &paymentDate1,
			PaymentRef:        "PAY-2024-003",
		},
	}

	for i := range orders {
		if err := db.Save(&orders[i]).Error; err != nil {
			return fmt.Errorf("failed to seed order: %w", err)
		}
	}

	// Create order items
	orderItems := []OrderItem{
		{OrderID: 15, ProductID: 1, Quantity: 1, Price: 13999.99},
		{OrderID: 15, ProductID: 3, Quantity: 2, Price: 2499.99},
		{OrderID: 16, ProductID: 2, Quantity: 1, Price: 32999.99},
		{OrderID: 17, ProductID: 3, Quantity: 1, Price: 2499.99},
		{OrderID: 18, ProductID: 4, Quantity: 1, Price: 4499.99},
		{OrderID: 19, ProductID: 5, Quantity: 1, Price: 799.99},
	}

	for i := range orderItems {
		if err := db.Save(&orderItems[i]).Error; err != nil {
			return fmt.Errorf("failed to seed order item: %w", err)
		}
	}

	fmt.Println("✓ Orders seeded")
	return nil
}

func seedAddons(db *gorm.DB) error {
	fmt.Println("🔌 Seeding addons...")

	features1, _ := json.Marshal([]string{"AI-powered recommendations", "Automated responses", "24/7 chatbot support"})
	features2, _ := json.Marshal([]string{"Multi-channel integration", "WhatsApp Business API", "Facebook Messenger", "Instagram DM"})
	features3, _ := json.Marshal([]string{"Advanced analytics", "Custom reports", "Data export", "Real-time dashboards"})

	addons := []Addon{
		{
			Title:        "AI Assistant Pro",
			Description:  "Advanced AI-powered assistant for customer support",
			Logo:         "https://picsum.photos/100/100?random=10",
			Photo:        "https://picsum.photos/400/300?random=11",
			Category:     "AI",
			PricingType:  "time",
			BasePrice:    499.99,
			Currency:     "EGP",
			BillingCycle: 30,
			Features:     string(features1),
			IsActive:     true,
		},
		{
			Title:        "Multi-Channel Integration",
			Description:  "Connect with your customers across all platforms",
			Logo:         "https://picsum.photos/100/100?random=12",
			Photo:        "https://picsum.photos/400/300?random=13",
			Category:     "Integration",
			PricingType:  "time",
			BasePrice:    299.99,
			Currency:     "EGP",
			BillingCycle: 30,
			Features:     string(features2),
			IsActive:     true,
		},
		{
			Title:       "Analytics Dashboard",
			Description: "Comprehensive analytics and reporting suite",
			Logo:        "https://picsum.photos/100/100?random=14",
			Photo:       "https://picsum.photos/400/300?random=15",
			Category:    "Analytics",
			PricingType: "usage",
			BasePrice:   0.99,
			Currency:    "EGP",
			UsageUnit:   "reports",
			Features:    string(features3),
			IsActive:    true,
		},
		{
			Title:       "Email Marketing Suite",
			Description: "Professional email campaigns and automation",
			Logo:        "https://picsum.photos/100/100?random=16",
			Photo:       "https://picsum.photos/400/300?random=17",
			Category:    "Marketing",
			PricingType: "usage",
			BasePrice:   0.05,
			Currency:    "EGP",
			UsageUnit:   "emails",
			Features:    string(features1),
			IsActive:    true,
		},
	}

	for i := range addons {
		var existing Addon
		result := db.Where("title = ?", addons[i].Title).First(&existing)
		if result.Error != nil {
			// Create if doesn't exist
			if err := db.Save(&addons[i]).Error; err != nil {
				return fmt.Errorf("failed to seed addon: %w", err)
			}
		}
	}

	// Create addon pricing tiers
	tiers := []AddonPricingTier{
		{AddonID: 1, MinQuantity: 3, MaxQuantity: 6, DiscountType: "percentage", DiscountValue: 10, FinalPrice: 1349.97, Description: "3 months - 10% off"},
		{AddonID: 1, MinQuantity: 6, MaxQuantity: 12, DiscountType: "percentage", DiscountValue: 15, FinalPrice: 2549.94, Description: "6 months - 15% off"},
		{AddonID: 2, MinQuantity: 3, MaxQuantity: 6, DiscountType: "percentage", DiscountValue: 10, FinalPrice: 809.97, Description: "3 months - 10% off"},
		{AddonID: 2, MinQuantity: 6, MaxQuantity: 12, DiscountType: "percentage", DiscountValue: 20, FinalPrice: 1439.95, Description: "6 months - 20% off"},
	}

	for i := range tiers {
		if err := db.Save(&tiers[i]).Error; err != nil {
			return fmt.Errorf("failed to seed pricing tier: %w", err)
		}
	}

	fmt.Println("✓ Addons seeded")
	return nil
}

func seedAddonSubscriptions(db *gorm.DB) error {
	fmt.Println("💳 Seeding addon subscriptions...")

	now := time.Now()
	endDate1 := now.Add(30 * 24 * time.Hour)
	endDate2 := now.Add(60 * 24 * time.Hour)
	usageLimit := 1000

	subscriptions := []UserAddonSubscription{
		{
			UserID:     2,
			AddonID:    1,
			Status:     AddonSubscriptionStatus_ACTIVE,
			Quantity:   1,
			TotalPrice: 499.99,
			StartDate:  now,
			EndDate:    &endDate1,
			AutoRenew:  true,
			PaymentID:  new(uint),
		},
		{
			UserID:     2,
			AddonID:    2,
			Status:     AddonSubscriptionStatus_ACTIVE,
			Quantity:   3,
			TotalPrice: 809.97,
			StartDate:  now,
			EndDate:    &endDate2,
			AutoRenew:  false,
			PaymentID:  new(uint),
		},
		{
			UserID:     2,
			AddonID:    3,
			Status:     AddonSubscriptionStatus_ACTIVE,
			Quantity:   500,
			TotalPrice: 495.00,
			StartDate:  now,
			UsageLimit: &usageLimit,
			UsageCount: 234,
			AutoRenew:  true,
		},
		{
			UserID:     3,
			AddonID:    1,
			Status:     AddonSubscriptionStatus_PENDING,
			Quantity:   1,
			TotalPrice: 499.99,
			StartDate:  now,
			EndDate:    &endDate1,
			AutoRenew:  false,
		},
	}

	*subscriptions[0].PaymentID = 20
	*subscriptions[1].PaymentID = 21

	for i := range subscriptions {
		if err := db.Save(&subscriptions[i]).Error; err != nil {
			return fmt.Errorf("failed to seed addon subscription: %w", err)
		}
	}

	fmt.Println("✓ Addon subscriptions seeded")
	return nil
}

func seedPayments(db *gorm.DB) error {
	fmt.Println("💰 Seeding payments...")

	now := time.Now()
	paidAt1 := now.Add(-2 * 24 * time.Hour)
	paidAt2 := now.Add(-5 * 24 * time.Hour)
	expiresAt := now.Add(24 * time.Hour)

	payments := []Payment{
		{
			UserID:          2,
			PackageID:       3,
			Amount:          2499.99,
			Currency:        "EGP",
			PaymentMethod:   "fawry",
			PaymentStatus:   PaymentStatus_PAID,
			TransactionID:   "TXN-FAW-001",
			ReferenceNumber: "FAW-REF-123425",
			PaidAt:          &paidAt1,
		},
		{
			UserID:        2,
			PackageID:     2,
			Amount:        899.99,
			Currency:      "EGP",
			PaymentMethod: "paymob",
			PaymentStatus: PaymentStatus_PAID,
			TransactionID: "TXN-PAY-002",
			PaymobOrderID: "PAYMOB-ORD-678950",
			PaidAt:        &paidAt2,
		},
		{
			UserID:          3,
			PackageID:       1,
			Amount:          299.99,
			Currency:        "EGP",
			PaymentMethod:   "fawry",
			PaymentStatus:   PaymentStatus_PENDING,
			ReferenceNumber: "FAW-REF-5435210",
			ExpiresAt:       &expiresAt,
		},
		{
			UserID:        3,
			PackageID:     2,
			Amount:        899.99,
			Currency:      "EGP",
			PaymentMethod: "paymob",
			PaymentStatus: PaymentStatus_FAILED,
			TransactionID: "TXN-PAY-003",
			PaymobOrderID: "PAYMOB-ORD-1115110",
		},
	}

	for i := range payments {
		if err := db.Save(&payments[i]).Error; err != nil {
			return fmt.Errorf("failed to seed payment: %w", err)
		}
	}

	fmt.Println("✓ Payments seeded")
	return nil
}

func seedTodos(db *gorm.DB) error {
	fmt.Println("✅ Seeding todos...")

	now := time.Now()
	dueDate1 := now.Add(24 * time.Hour)
	dueDate2 := now.Add(48 * time.Hour)
	dueDate3 := now.Add(72 * time.Hour)
	completedAt := now.Add(-24 * time.Hour)

	todos := []Todo{
		{
			Title:       "Review Q4 sales reports",
			Description: "Analyze sales performance and prepare summary",
			IsCompleted: true,
			Priority:    "high",
			DueDate:     &dueDate1,
			CompletedAt: &completedAt,
			UserID:      2,
		},
		{
			Title:       "Update product inventory",
			Description: "Check stock levels and reorder items",
			IsCompleted: false,
			Priority:    "urgent",
			DueDate:     &dueDate1,
			UserID:      2,
		},
		{
			Title:       "Client meeting preparation",
			Description: "Prepare presentation for Acme Corporation",
			IsCompleted: false,
			Priority:    "high",
			DueDate:     &dueDate2,
			UserID:      2,
		},
		{
			Title:       "Review marketing campaign",
			Description: "Check email campaign performance metrics",
			IsCompleted: false,
			Priority:    "medium",
			DueDate:     &dueDate3,
			UserID:      2,
		},
		{
			Title:       "Team sync meeting",
			Description: "Weekly team standup",
			IsCompleted: true,
			Priority:    "low",
			CompletedAt: &completedAt,
			UserID:      3,
		},
	}

	for i := range todos {
		if err := db.Save(&todos[i]).Error; err != nil {
			return fmt.Errorf("failed to seed todo: %w", err)
		}
	}

	fmt.Println("✓ Todos seeded")
	return nil
}

func seedCalendarEvents(db *gorm.DB) error {
	fmt.Println("📅 Seeding calendar events...")

	now := time.Now()
	event1Start := now.Add(2 * time.Hour)
	event1End := event1Start.Add(1 * time.Hour)
	event2Start := now.Add(24 * time.Hour)
	event2End := event2Start.Add(2 * time.Hour)
	event3Start := now.Add(48 * time.Hour)
	event3End := event3Start.Add(30 * time.Minute)

	events := []CalendarEvent{
		{
			UserID:       2,
			Title:        "Client Strategy Meeting",
			Description:  "Quarterly review with Acme Corporation",
			Location:     "Conference Room A",
			StartTime:    event1Start,
			EndTime:      event1End,
			AllDay:       false,
			EventType:    "meeting",
			Color:        "#3B82F6",
			IsPublic:     false,
			RemindBefore: 30,
			Status:       EventStatus_SCHEDULED,
		},
		{
			UserID:       2,
			Title:        "Product Launch Webinar",
			Description:  "Launch event for new product line",
			Location:     "Online - Zoom",
			StartTime:    event2Start,
			EndTime:      event2End,
			AllDay:       false,
			EventType:    "public",
			Color:        "#10B981",
			IsPublic:     true,
			RemindBefore: 60,
			Status:       EventStatus_SCHEDULED,
		},
		{
			UserID:         2,
			Title:          "Team Standup",
			Description:    "Daily team sync",
			Location:       "Virtual",
			StartTime:      event3Start,
			EndTime:        event3End,
			AllDay:         false,
			EventType:      "meeting",
			Color:          "#8B5CF6",
			IsPublic:       false,
			Recurring:      true,
			RecurrenceRule: "FREQ=DAILY;COUNT=30",
			RemindBefore:   15,
			Status:         EventStatus_SCHEDULED,
		},
		{
			UserID:      3,
			Title:       "Company Holiday",
			Description: "National Holiday",
			StartTime:   now.Add(72 * time.Hour),
			EndTime:     now.Add(96 * time.Hour),
			AllDay:      true,
			EventType:   "reminder",
			Color:       "#EF4444",
			IsPublic:    true,
			Status:      EventStatus_SCHEDULED,
		},
	}

	for i := range events {
		if err := db.Save(&events[i]).Error; err != nil {
			return fmt.Errorf("failed to seed calendar event: %w", err)
		}
	}

	// Add attendees
	attendees := []EventAttendee{
		{EventID: 1, UserID: new(uint), Email: "client@acme.com", Name: "John Acme", ResponseStatus: "accepted", NotificationSent: true},
		{EventID: 1, UserID: new(uint), Email: "sarah.j@example.com", Name: "Sarah Johnson", ResponseStatus: "pending"},
		{EventID: 2, UserID: new(uint), Email: "marketing@example.com", Name: "Marketing Team", ResponseStatus: "accepted"},
	}

	*attendees[0].UserID = 2
	*attendees[2].UserID = 3

	for i := range attendees {
		if err := db.Save(&attendees[i]).Error; err != nil {
			return fmt.Errorf("failed to seed event attendee: %w", err)
		}
	}

	fmt.Println("✓ Calendar events seeded")
	return nil
}

func seedNotifications(db *gorm.DB) error {
	fmt.Println("🔔 Seeding notifications...")

	now := time.Now()
	readAt := now.Add(-1 * time.Hour)

	notifications := []Notification{
		{
			UserID:  2,
			Title:   "New Order Received",
			Message: "You have received a new order #1001 from Acme Corporation",
			Type:    "success",
			IsRead:  true,
			ReadAt:  &readAt,
			Link:    "/orders/1001",
		},
		{
			UserID:  2,
			Title:   "Payment Successful",
			Message: "Your payment of 2499.99 EGP has been processed successfully",
			Type:    "success",
			IsRead:  false,
			Link:    "/payments/1",
		},
		{
			UserID:  2,
			Title:   "Subscription Expiring Soon",
			Message: "Your Professional package will expire in 7 days",
			Type:    "warning",
			IsRead:  false,
			Link:    "/subscriptions",
		},
		{
			UserID:  2,
			Title:   "New Feature Available",
			Message: "Check out our new AI Assistant Pro addon!",
			Type:    "info",
			IsRead:  false,
			Link:    "/addons/1",
		},
		{
			UserID:  3,
			Title:   "Account Verification Required",
			Message: "Please verify your email address to unlock all features",
			Type:    "warning",
			IsRead:  false,
			Link:    "/settings/verification",
		},
	}

	for i := range notifications {
		if err := db.Save(&notifications[i]).Error; err != nil {
			return fmt.Errorf("failed to seed notification: %w", err)
		}
	}

	fmt.Println("✓ Notifications seeded")
	return nil
}

func seedBlogs(db *gorm.DB) error {
	fmt.Println("📝 Seeding blogs...")

	now := time.Now()
	publishedAt := now.Add(-48 * time.Hour)

	photos1, _ := json.Marshal([]string{"https://picsum.photos/800/600?random=20", "https://picsum.photos/800/600?random=21"})
	photos2, _ := json.Marshal([]string{"https://picsum.photos/800/600?random=22"})

	blogs := []Blog{
		{
			Title:       "10 Tips for Growing Your E-commerce Business",
			Content:     "In this comprehensive guide, we'll explore proven strategies to scale your online business...",
			Summary:     "Learn essential tips for e-commerce growth",
			Slug:        "10-tips-ecommerce-growth",
			AuthorID:    1,
			Photos:      string(photos1),
			IsPublished: true,
			PublishedAt: &publishedAt,
		},
		{
			Title:       "Understanding Customer Behavior in 2024",
			Content:     "Customer preferences are constantly evolving. Here's what you need to know...",
			Summary:     "Insights into modern customer behavior",
			Slug:        "customer-behavior-2024",
			AuthorID:    1,
			Photos:      string(photos2),
			IsPublished: true,
			PublishedAt: &publishedAt,
		},
		{
			Title:       "The Future of AI in Business",
			Content:     "Artificial intelligence is transforming how businesses operate. Discover the latest trends...",
			Summary:     "AI trends and business applications",
			Slug:        "future-ai-business",
			AuthorID:    2,
			Photos:      string(photos1),
			IsPublished: false,
		},
	}

	for i := range blogs {
		var existing Blog
		result := db.Where("slug = ?", blogs[i].Slug).First(&existing)
		if result.Error != nil {
			// Create if doesn't exist
			if err := db.Save(&blogs[i]).Error; err != nil {
				return fmt.Errorf("failed to seed blog: %w", err)
			}
		}
	}

	fmt.Println("✓ Blogs seeded")
	return nil
}

func seedNewsletters(db *gorm.DB) error {
	fmt.Println("📧 Seeding newsletters...")

	now := time.Now()
	unsubscribedAt := now.Add(-10 * 24 * time.Hour)

	newsletters := []Newsletter{
		{Email: "subscriber1@example.com", IsActive: true, SubscribedAt: now},
		{Email: "subscriber2@example.com", IsActive: true, SubscribedAt: now},
		{Email: "subscriber3@example.com", IsActive: false, SubscribedAt: now.Add(-30 * 24 * time.Hour), UnsubscribedAt: &unsubscribedAt},
		{Email: "subscriber4@example.com", IsActive: true, SubscribedAt: now},
	}

	for i := range newsletters {
		var existing Newsletter
		result := db.Where("email = ?", newsletters[i].Email).First(&existing)
		if result.Error != nil {
			// Create if doesn't exist
			if err := db.Save(&newsletters[i]).Error; err != nil {
				return fmt.Errorf("failed to seed newsletter: %w", err)
			}
		}
	}

	fmt.Println("✓ Newsletters seeded")
	return nil
}

func seedContacts(db *gorm.DB) error {
	fmt.Println("📞 Seeding contacts...")

	contacts := []Contact{
		{
			Name:    "David Brown",
			Email:   "david.b@example.com",
			Message: "I'm interested in the Enterprise package. Can you provide more details?",
			IsRead:  true,
			Replied: true,
		},
		{
			Name:    "Lisa Wilson",
			Email:   "lisa.w@example.com",
			Message: "Having trouble with payment integration. Please help.",
			IsRead:  true,
			Replied: false,
		},
		{
			Name:    "Robert Taylor",
			Email:   "robert.t@example.com",
			Message: "Great platform! Just wanted to share my positive experience.",
			IsRead:  false,
			Replied: false,
		},
	}

	for i := range contacts {
		if err := db.Save(&contacts[i]).Error; err != nil {
			return fmt.Errorf("failed to seed contact: %w", err)
		}
	}

	fmt.Println("✓ Contacts seeded")
	return nil
}

func seedReceipts(db *gorm.DB) error {
	fmt.Println("🧾 Seeding receipts...")

	now := time.Now()
	generatedAt := now.Add(-1 * time.Hour)

	companyInfo, _ := json.Marshal(map[string]string{
		"name":    "Hamber Hub Ltd",
		"address": "123 Business Street, Cairo, Egypt",
		"tax_id":  "TAX-123456789",
		"phone":   "+20-2-1234-5678",
	})

	receipts := []OrderReceipt{
		{
			OrderID:         15,
			ReceiptNumber:   "RCP-2024-001",
			PDFPath:         "/receipts/RCP-2024-001.pdf",
			GeneratedAt:     &generatedAt,
			TemplateVersion: "v1",
			CompanyInfo:     string(companyInfo),
			Notes:           "Thank you for your business!",
		},
		{
			OrderID:         16,
			ReceiptNumber:   "RCP-2024-002",
			PDFPath:         "/receipts/RCP-2024-002.pdf",
			GeneratedAt:     &generatedAt,
			TemplateVersion: "v1",
			CompanyInfo:     string(companyInfo),
		},
	}

	for i := range receipts {
		var existing OrderReceipt
		result := db.Where("receipt_number = ?", receipts[i].ReceiptNumber).First(&existing)
		if result.Error != nil {
			// Create if doesn't exist
			if err := db.Save(&receipts[i]).Error; err != nil {
				return fmt.Errorf("failed to seed receipt: %w", err)
			}
		}
	}

	fmt.Println("✓ Receipts seeded")
	return nil
}
