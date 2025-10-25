// controllers/swagger_annotations.go
// Add these Swagger annotations to your controllers

package controllers

// Register godoc
// @Summary      Register a new user
// @Description  Register a new user account with email and password
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "Registration details"
// @Success      201 {object} AuthResponse "User created successfully"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Failure      409 {object} map[string]interface{} "Email or username already exists"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /auth/register [post]

// Login godoc
// @Summary      User login
// @Description  Authenticate user with email and password
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login credentials"
// @Success      200 {object} map[string]interface{} "List of packages"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /packages [get]

// GetPackage godoc
// @Summary      Get package by ID
// @Description  Get details of a specific package
// @Tags         Packages
// @Accept       json
// @Produce      json
// @Param        id path int true "Package ID"
// @Success      200 {object} dbmodels.Package "Package details"
// @Failure      404 {object} map[string]interface{} "Package not found"
// @Router       /packages/{id} [get]

// CreateProduct godoc
// @Summary      Create a new product
// @Description  Create a new product with photos (base64 encoded)
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body CreateProductRequest true "Product details"
// @Success      201 {object} map[string]interface{} "Product created"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /products [post]

// GetProducts godoc
// @Summary      Get products list
// @Description  Get paginated list of products
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(10)
// @Param        category query string false "Filter by category"
// @Success      200 {object} map[string]interface{} "Products list"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /products [get]

// GetProduct godoc
// @Summary      Get product by ID
// @Description  Get details of a specific product
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "Product ID"
// @Success      200 {object} ProductResponse "Product details"
// @Failure      404 {object} map[string]interface{} "Product not found"
// @Router       /products/{id} [get]

// UpdateProduct godoc
// @Summary      Update product
// @Description  Update an existing product
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "Product ID"
// @Param        request body UpdateProductRequest true "Updated product details"
// @Success      200 {object} map[string]interface{} "Product updated"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Router       /products/{id} [put]

// DeleteProduct godoc
// @Summary      Delete product
// @Description  Soft delete a product
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "Product ID"
// @Success      200 {object} map[string]interface{} "Product deleted"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Failure      404 {object} map[string]interface{} "Product not found"
// @Router       /products/{id} [delete]

// CreateOrder godoc
// @Summary      Create a new order
// @Description  Create a new order for products
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body CreateOrderRequest true "Order details"
// @Success      201 {object} map[string]interface{} "Order created"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /orders [post]

// GetOrders godoc
// @Summary      Get orders list
// @Description  Get paginated list of user orders
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(10)
// @Success      200 {object} map[string]interface{} "Orders list"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /orders [get]

// GetOrder godoc
// @Summary      Get order by ID
// @Description  Get details of a specific order
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "Order ID"
// @Success      200 {object} map[string]interface{} "Order details"
// @Failure      403 {object} map[string]interface{} "Forbidden"
// @Failure      404 {object} map[string]interface{} "Order not found"
// @Router       /orders/{id} [get]

// UpdateOrderStatus godoc
// @Summary      Update order status
// @Description  Update the status of an order
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "Order ID"
// @Param        request body map[string]string true "Status update"
// @Success      200 {object} map[string]interface{} "Status updated"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Router       /orders/{id}/status [patch]

// CreateTodo godoc
// @Summary      Create a new todo
// @Description  Create a new todo task
// @Tags         Todos
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body CreateTodoRequest true "Todo details"
// @Success      201 {object} map[string]interface{} "Todo created"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Router       /todos [post]

// GetTodos godoc
// @Summary      Get todos list
// @Description  Get paginated list of user todos
// @Tags         Todos
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(10)
// @Param        completed query boolean false "Filter by completion status"
// @Success      200 {object} map[string]interface{} "Todos list"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /todos [get]

// CreateBlog godoc
// @Summary      Create a new blog post
// @Description  Create a new blog post with photos
// @Tags         Blogs
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body CreateBlogRequest true "Blog details"
// @Success      201 {object} map[string]interface{} "Blog created"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Failure      409 {object} map[string]interface{} "Slug already exists"
// @Router       /blogs [post]

// GetBlogs godoc
// @Summary      Get blogs list
// @Description  Get paginated list of published blogs
// @Tags         Blogs
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(10)
// @Success      200 {object} map[string]interface{} "Blogs list"
// @Router       /blogs [get]

// RequestPackageChange godoc
// @Summary      Request package change
// @Description  Request to upgrade or downgrade package
// @Tags         Payments
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body ChangePackageRequest true "Package change request"
// @Success      200 {object} ChangePackageResponse "Change request created"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Router       /payment/change-package [post]

// GetUserNotifications godoc
// @Summary      Get user notifications
// @Description  Get paginated list of user notifications
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(20)
// @Param        unread_only query boolean false "Show only unread" default(false)
// @Success      200 {object} map[string]interface{} "Notifications list"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /notifications [get]

// MarkNotificationAsRead godoc
// @Summary      Mark notification as read
// @Description  Mark a specific notification as read
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path int true "Notification ID"
// @Success      200 {object} map[string]interface{} "Notification marked as read"
// @Failure      404 {object} map[string]interface{} "Notification not found"
// @Router       /notifications/{id}/read [patch]

// GetUnreadCount godoc
// @Summary      Get unread notifications count
// @Description  Get count of unread notifications for current user
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} map[string]interface{} "Unread count"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /notifications/unread-count [get]

// GetAddons godoc
// @Summary      Get add-ons list
// @Description  Get paginated list of available add-ons
// @Tags         Add-ons
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(20)
// @Param        category query string false "Filter by category"
// @Param        active query boolean false "Filter by active status"
// @Success      200 {object} map[string]interface{} "Add-ons list"
// @Router       /addons [get]

// SubscribeToAddon godoc
// @Summary      Subscribe to an add-on
// @Description  Subscribe to an add-on service
// @Tags         Add-on Subscriptions
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body SubscribeAddonRequest true "Subscription details"
// @Success      201 {object} map[string]interface{} "Subscription created"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Router       /subscriptions [post]

// GetUserSubscriptions godoc
// @Summary      Get user subscriptions
// @Description  Get all add-on subscriptions for current user
// @Tags         Add-on Subscriptions
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        status query string false "Filter by status"
// @Success      200 {object} map[string]interface{} "Subscriptions list"
// @Router       /subscriptions [get]

// CreateCalendarEvent godoc
// @Summary      Create calendar event
// @Description  Create a new calendar event
// @Tags         Calendar
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body CreateEventRequest true "Event details"
// @Success      201 {object} map[string]interface{} "Event created"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Router       /calendar/events [post]

// GetUserEvents godoc
// @Summary      Get user events
// @Description  Get calendar events for a specific month
// @Tags         Calendar
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        year query int false "Year" default(2025)
// @Param        month query int false "Month (1-12)" default(10)
// @Param        include_public query boolean false "Include public events" default(true)
// @Success      200 {object} map[string]interface{} "Events list"
// @Router       /calendar/events [get]

// GenerateOrderReceipt godoc
// @Summary      Generate order receipt
// @Description  Generate PDF receipt for an order
// @Tags         Receipts
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        order_id path int true "Order ID"
// @Param        request body map[string]interface{} false "Company info"
// @Success      201 {object} map[string]interface{} "Receipt generated"
// @Failure      404 {object} map[string]interface{} "Order not found"
// @Router       /receipts/order/{order_id} [post]

// DownloadReceipt godoc
// @Summary      Download receipt PDF
// @Description  Download receipt as PDF file
// @Tags         Receipts
// @Accept       json
// @Produce      application/pdf
// @Security     Bearer
// @Param        order_id path int true "Order ID"
// @Success      200 {file} file "PDF file"
// @Failure      404 {object} map[string]interface{} "Receipt not found"
// @Router       /receipts/order/{order_id}/download [get]{object} AuthResponse "Login successful"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Failure      401 {object} map[string]interface{} "Invalid credentials"
// @Router       /auth/login [post]

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Get a new access token using refresh token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body RefreshTokenRequest true "Refresh token"
// @Success      200 {object} map[string]interface{} "New tokens generated"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Failure      401 {object} map[string]interface{} "Invalid refresh token"
// @Router       /auth/refresh [post]

// GetProfile godoc
// @Summary      Get user profile
// @Description  Get current user's profile information
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} dbmodels.User "User profile"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "User not found"
// @Router       /profile [get]

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Update current user's profile information
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body map[string]interface{} true "Profile update data"
// @Success      200 {object} dbmodels.User "Updated profile"
// @Failure      400 {object} map[string]interface{} "Invalid request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /profile [put]

// GetAllPackages godoc
// @Summary      Get all packages
// @Description  Get list of all available packages
// @Tags         Packages
// @Accept       json
// @Produce      json
// @Success      200
