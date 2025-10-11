# API Documentation for Frontend

## Base URL
```
https://test.hamber-hub.com/api
```

## Table of Contents
- [Authentication](#authentication)
- [User Management](#user-management)
- [Package Management](#package-management)
- [Payment & Billing](#payment--billing)
- [Profile Management](#profile-management)
- [Blog Management](#blog-management)
- [Newsletter](#newsletter)
- [Contact](#contact)
- [Products](#products)
- [Orders](#orders)
- [Todos](#todos)
- [Admin Routes](#admin-routes)

---

## Authentication

### Register New User
**Endpoint:** `POST /auth/register`  
**Authentication:** None  
**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123",
  "subdomain": "johndoe",
  "package_id": 1
}
```
**Response:** `201 Created`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "ID": 1,
    "Name": "John Doe",
    "Email": "john@example.com",
    "PackageID": 1
  }
}
```

### Login
**Endpoint:** `POST /auth/login`  
**Authentication:** None  
**Request Body:**
```json
{
  "email": "john@example.com",
  "password": "password123"
}
```
**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "ID": 1,
    "Name": "John Doe",
    "Email": "john@example.com"
  }
}
```

### Refresh Token
**Endpoint:** `POST /auth/refresh`  
**Authentication:** None  
**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```
**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### Forgot Password
**Endpoint:** `POST /auth/forgot-password`  
**Authentication:** None  
**Request Body:**
```json
{
  "email": "john@example.com"
}
```
**Response:** `200 OK`
```json
{
  "message": "Password reset code sent to your email"
}
```

### Reset Password
**Endpoint:** `POST /auth/reset-password`  
**Authentication:** None  
**Request Body:**
```json
{
  "email": "john@example.com",
  "code": "123456",
  "new_password": "newpassword123"
}
```
**Response:** `200 OK`
```json
{
  "message": "Password reset successfully"
}
```

---

## OAuth Authentication

### Google Login
**Endpoint:** `GET /auth/oauth/google`  
**Authentication:** None  
**Description:** Redirects to Google OAuth login page

### Google Callback
**Endpoint:** `GET /auth/oauth/google/callback`  
**Authentication:** None  
**Description:** Called by Google after authentication

### Facebook Login
**Endpoint:** `GET /auth/oauth/facebook`  
**Authentication:** None  
**Description:** Redirects to Facebook OAuth login page

### Facebook Callback
**Endpoint:** `GET /auth/oauth/facebook/callback`  
**Authentication:** None  
**Description:** Called by Facebook after authentication

### Apple Login
**Endpoint:** `GET /auth/oauth/apple`  
**Authentication:** None  
**Description:** Redirects to Apple OAuth login page

### Apple Callback
**Endpoint:** `GET /auth/oauth/apple/callback`  
**Authentication:** None  
**Description:** Called by Apple after authentication

---

## Email Verification

### Send Verification Email
**Endpoint:** `POST /verify/send-email`  
**Authentication:** None  
**Request Body:**
```json
{
  "email": "john@example.com"
}
```
**Response:** `200 OK`
```json
{
  "message": "Verification code sent to your email"
}
```

### Verify Email
**Endpoint:** `POST /verify/email`  
**Authentication:** None  
**Request Body:**
```json
{
  "email": "john@example.com",
  "code": "123456"
}
```
**Response:** `200 OK`
```json
{
  "message": "Email verified successfully"
}
```

---

## User Management

### Get Current User Profile
**Endpoint:** `GET /profile`  
**Authentication:** Required (Bearer Token)  
**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```
**Response:** `200 OK`
```json
{
  "ID": 1,
  "Name": "John Doe",
  "Email": "john@example.com",
  "Phone": "+201234567890",
  "Subdomain": "johndoe",
  "PackageID": 1,
  "Avatar": "https://...",
  "Bio": "Software developer",
  "Website": "https://johndoe.com",
  "Location": "Cairo, Egypt"
}
```

### Update Profile
**Endpoint:** `PUT /profile`  
**Authentication:** Required  
**Request Body:**
```json
{
  "name": "John Updated",
  "bio": "Senior Software Developer",
  "website": "https://newsite.com",
  "location": "Alexandria, Egypt"
}
```
**Response:** `200 OK`
```json
{
  "ID": 1,
  "Name": "John Updated",
  "Bio": "Senior Software Developer",
  "Website": "https://newsite.com",
  "Location": "Alexandria, Egypt"
}
```

### Get User Permissions
**Endpoint:** `GET /permissions`  
**Authentication:** Required  
**Response:** `200 OK`
```json
{
  "user_permissions": {
    "user_id": 1,
    "email": "john@example.com",
    "role": "user",
    "permissions": [
      {
        "ID": 1,
        "Name": "CREATE_POST"
      },
      {
        "ID": 2,
        "Name": "EDIT_POST"
      }
    ]
  },
  "jwt_permissions": ["CREATE_POST", "EDIT_POST"],
  "message": "Permissions retrieved successfully"
}
```

---

## Package Management

### Get All Packages
**Endpoint:** `GET /packages`  
**Authentication:** None  
**Response:** `200 OK`
```json
{
  "packages": [
    {
      "id": 1,
      "name": "Free Plan",
      "price": 0,
      "duration": 30,
      "benefits": "[\"10 GB Storage\", \"Basic Support\"]",
      "description": "Perfect for getting started",
      "is_active": true,
      "price_per_client": false
    },
    {
      "id": 2,
      "name": "Premium Plan",
      "price": 299.99,
      "duration": 30,
      "benefits": "[\"100 GB Storage\", \"Priority Support\", \"Custom Domain\"]",
      "description": "For growing businesses",
      "is_active": true,
      "price_per_client": false
    }
  ]
}
```

### Get Single Package
**Endpoint:** `GET /packages/:id`  
**Authentication:** None  
**Response:** `200 OK`
```json
{
  "package": {
    "id": 2,
    "name": "Premium Plan",
    "price": 299.99,
    "duration": 30,
    "benefits": "[\"100 GB Storage\", \"Priority Support\"]",
    "description": "For growing businesses",
    "is_active": true
  }
}
```

---

## Payment & Billing

### Request Package Change
**Endpoint:** `POST /payment/change-package`  
**Authentication:** Required  
**Request Body:**
```json
{
  "new_package_id": 2,
  "payment_method": "fawry",
  "reason": "Upgrading to premium for more storage"
}
```
**Response (Fawry):** `200 OK`
```json
{
  "package_change_id": 1,
  "payment_id": 1,
  "reference_number": "1234567890",
  "message": "Please pay at any Fawry location using reference number: 1234567890",
  "amount": 299.99,
  "expires_at": "2025-10-12T10:00:00Z"
}
```
**Response (Paymob):** `200 OK`
```json
{
  "package_change_id": 1,
  "payment_id": 1,
  "payment_url": "https://accept.paymob.com/api/acceptance/iframes/123?payment_token=xyz",
  "message": "Please complete payment using the provided URL",
  "amount": 299.99,
  "expires_at": "2025-10-12T10:00:00Z"
}
```

### Get Payment Status
**Endpoint:** `GET /payment/status/:id`  
**Authentication:** Required  
**Response:** `200 OK`
```json
{
  "id": 1,
  "user_id": 1,
  "package_id": 2,
  "amount": 299.99,
  "currency": "EGP",
  "payment_method": "fawry",
  "payment_status": 1,
  "reference_number": "1234567890",
  "transaction_id": "TXN123456",
  "paid_at": "2025-10-11T15:30:00Z",
  "created_at": "2025-10-11T10:00:00Z"
}
```

### Get Payment History
**Endpoint:** `GET /payment/history?page=1&limit=20`  
**Authentication:** Required  
**Query Parameters:**
- `page` (optional, default: 1)
- `limit` (optional, default: 20, max: 100)

**Response:** `200 OK`
```json
{
  "payments": [
    {
      "id": 1,
      "package_id": 2,
      "amount": 299.99,
      "payment_method": "fawry",
      "payment_status": 1,
      "created_at": "2025-10-11T10:00:00Z"
    }
  ],
  "total": 5,
  "page": 1,
  "limit": 20
}
```

### Get Package Change History
**Endpoint:** `GET /payment/package-changes?page=1&limit=20`  
**Authentication:** Required  
**Response:** `200 OK`
```json
{
  "changes": [
    {
      "id": 1,
      "old_package_id": 1,
      "new_package_id": 2,
      "status": 3,
      "approved_at": "2025-10-11T15:30:00Z",
      "created_at": "2025-10-11T10:00:00Z"
    }
  ],
  "total": 3,
  "page": 1,
  "limit": 20
}
```

---

## Profile & Photos

### Upload Avatar Photo
**Endpoint:** `POST /photos/avatar`  
**Authentication:** Required  
**Content-Type:** `multipart/form-data`  
**Request Body:**
```
photo: [file]
```
**Response:** `200 OK`
```json
{
  "message": "Avatar uploaded successfully",
  "avatar_url": "https://minio.example.com/bucket/users/1/avatar.jpg"
}
```

### Get Photo Presigned URL
**Endpoint:** `GET /photos/presigned-url?object_name=users/1/avatar.jpg`  
**Authentication:** Required  
**Query Parameters:**
- `object_name` (required)

**Response:** `200 OK`
```json
{
  "url": "https://minio.example.com/bucket/users/1/avatar.jpg?X-Amz-Signature=..."
}
```

---

## Blog Management

### Get All Blogs (Public)
**Endpoint:** `GET /blogs?page=1&limit=20`  
**Authentication:** None  
**Query Parameters:**
- `page` (optional, default: 1)
- `limit` (optional, default: 20)

**Response:** `200 OK`
```json
{
  "blogs": [
    {
      "id": 1,
      "title": "Getting Started with Go",
      "content": "Full article content...",
      "summary": "Learn Go basics",
      "slug": "getting-started-with-go",
      "author_id": 1,
      "author": {
        "ID": 1,
        "Name": "John Doe"
      },
      "photos": "[\"https://...\", \"https://...\"]",
      "is_published": true,
      "published_at": "2025-10-10T10:00:00Z",
      "created_at": "2025-10-09T10:00:00Z"
    }
  ],
  "total": 10,
  "page": 1,
  "limit": 20
}
```

### Get Single Blog
**Endpoint:** `GET /blogs/:id`  
**Authentication:** None  
**Response:** `200 OK`
```json
{
  "id": 1,
  "title": "Getting Started with Go",
  "content": "Full article content...",
  "slug": "getting-started-with-go",
  "author": {
    "ID": 1,
    "Name": "John Doe"
  },
  "photos": "[\"https://...\"]",
  "is_published": true,
  "published_at": "2025-10-10T10:00:00Z"
}
```

### Create Blog (Protected)
**Endpoint:** `POST /blogs`  
**Authentication:** Required  
**Request Body:**
```json
{
  "title": "My New Blog Post",
  "content": "Blog content here...",
  "summary": "Short summary",
  "slug": "my-new-blog-post",
  "is_published": false
}
```
**Response:** `201 Created`
```json
{
  "message": "Blog created successfully",
  "blog": {
    "id": 2,
    "title": "My New Blog Post",
    "slug": "my-new-blog-post"
  }
}
```

### Update Blog
**Endpoint:** `PUT /blogs/:id`  
**Authentication:** Required  
**Request Body:**
```json
{
  "title": "Updated Title",
  "content": "Updated content...",
  "is_published": true
}
```
**Response:** `200 OK`
```json
{
  "message": "Blog updated successfully",
  "blog": {
    "id": 2,
    "title": "Updated Title"
  }
}
```

### Delete Blog
**Endpoint:** `DELETE /blogs/:id`  
**Authentication:** Required  
**Response:** `200 OK`
```json
{
  "message": "Blog deleted successfully"
}
```

### Upload Blog Photos
**Endpoint:** `POST /blogs/:id/photos`  
**Authentication:** Required  
**Content-Type:** `multipart/form-data`  
**Request Body:**
```
photos: [files]
```
**Response:** `200 OK`
```json
{
  "message": "Photos uploaded successfully",
  "photo_urls": [
    "https://minio.example.com/bucket/blogs/2/photo1.jpg",
    "https://minio.example.com/bucket/blogs/2/photo2.jpg"
  ]
}
```

### Delete Blog Photo
**Endpoint:** `DELETE /blogs/:id/photos`  
**Authentication:** Required  
**Request Body:**
```json
{
  "photo_url": "https://minio.example.com/bucket/blogs/2/photo1.jpg"
}
```
**Response:** `200 OK`
```json
{
  "message": "Photo deleted successfully"
}
```

---

## Newsletter

### Subscribe to Newsletter
**Endpoint:** `POST /newsletter/subscribe`  
**Authentication:** None  
**Request Body:**
```json
{
  "email": "subscriber@example.com"
}
```
**Response:** `200 OK`
```json
{
  "message": "Successfully subscribed to newsletter"
}
```

### Unsubscribe from Newsletter
**Endpoint:** `POST /newsletter/unsubscribe`  
**Authentication:** None  
**Request Body:**
```json
{
  "email": "subscriber@example.com"
}
```
**Response:** `200 OK`
```json
{
  "message": "Successfully unsubscribed from newsletter"
}
```

---

## Contact

### Submit Contact Form
**Endpoint:** `POST /contact`  
**Authentication:** None  
**Request Body:**
```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "message": "I have a question about your services..."
}
```
**Response:** `200 OK`
```json
{
  "message": "Contact form submitted successfully"
}
```

---

## Products

### Create Product
**Endpoint:** `POST /products`  
**Authentication:** Required  
**Request Body:**
```json
{
  "name": "Product Name",
  "description": "Product description",
  "price": 99.99,
  "discount_price": 79.99,
  "quantity": 100,
  "sku": "PROD-001",
  "category": "Electronics",
  "brand": "Brand Name",
  "images": "[\"https://...\", \"https://...\"]",
  "weight": 1.5,
  "tags": "[\"tag1\", \"tag2\"]"
}
```
**Response:** `201 Created`
```json
{
  "message": "Product created successfully",
  "product": {
    "id": 1,
    "name": "Product Name",
    "sku": "PROD-001"
  }
}
```

### Get All Products
**Endpoint:** `GET /products?page=1&limit=20&category=Electronics`  
**Authentication:** Required  
**Query Parameters:**
- `page` (optional)
- `limit` (optional)
- `category` (optional)
- `is_active` (optional)

**Response:** `200 OK`
```json
{
  "products": [
    {
      "id": 1,
      "name": "Product Name",
      "price": 99.99,
      "discount_price": 79.99,
      "quantity": 100,
      "sku": "PROD-001",
      "category": "Electronics",
      "is_active": true
    }
  ],
  "total": 50,
  "page": 1,
  "limit": 20
}
```

### Get Single Product
**Endpoint:** `GET /products/:id`  
**Authentication:** Required  
**Response:** `200 OK`
```json
{
  "id": 1,
  "name": "Product Name",
  "description": "Product description",
  "price": 99.99,
  "quantity": 100,
  "images": "[\"https://...\"]"
}
```

### Update Product
**Endpoint:** `PUT /products/:id`  
**Authentication:** Required  
**Request Body:**
```json
{
  "name": "Updated Product Name",
  "price": 89.99,
  "quantity": 150
}
```
**Response:** `200 OK`
```json
{
  "message": "Product updated successfully"
}
```

### Delete Product
**Endpoint:** `DELETE /products/:id`  
**Authentication:** Required  
**Response:** `200 OK`
```json
{
  "message": "Product deleted successfully"
}
```

### Update Product Quantity
**Endpoint:** `PATCH /products/:id/quantity`  
**Authentication:** Required  
**Request Body:**
```json
{
  "quantity": 200
}
```
**Response:** `200 OK`
```json
{
  "message": "Product quantity updated successfully"
}
```

---

## Orders

### Create Order
**Endpoint:** `POST /orders`  
**Authentication:** Required  
**Request Body:**
```json
{
  "client_id": 1,
  "items": [
    {
      "product_id": 1,
      "quantity": 2
    },
    {
      "product_id": 2,
      "quantity": 1
    }
  ]
}
```
**Response:** `201 Created`
```json
{
  "message": "Order created successfully",
  "order": {
    "id": 1,
    "total": 199.98,
    "status": 0
  }
}
```

### Get All Orders
**Endpoint:** `GET /orders?page=1&limit=20&status=0`  
**Authentication:** Required  
**Query Parameters:**
- `page` (optional)
- `limit` (optional)
- `status` (optional): 0=PENDING, 1=SHIPPED, 2=DELIVERED, 3=CANCELED

**Response:** `200 OK`
```json
{
  "orders": [
    {
      "id": 1,
      "client_id": 1,
      "total": 199.98,
      "status": 0,
      "created_at": "2025-10-11T10:00:00Z"
    }
  ],
  "total": 10,
  "page": 1,
  "limit": 20
}
```

### Get Single Order
**Endpoint:** `GET /orders/:id`  
**Authentication:** Required  
**Response:** `200 OK`
```json
{
  "id": 1,
  "client": {
    "ID": 1,
    "Name": "Client Name"
  },
  "total": 199.98,
  "status": 0,
  "created_at": "2025-10-11T10:00:00Z"
}
```

### Update Order Status
**Endpoint:** `PATCH /orders/:id/status`  
**Authentication:** Required  
**Request Body:**
```json
{
  "status": 1
}
```
**Response:** `200 OK`
```json
{
  "message": "Order status updated successfully"
}
```

### Cancel Order
**Endpoint:** `PATCH /orders/:id/cancel`  
**Authentication:** Required  
**Response:** `200 OK`
```json
{
  "message": "Order cancelled successfully"
}
```

---

## Todos

### Create Todo
**Endpoint:** `POST /todos`  
**Authentication:** Required  
**Request Body:**
```json
{
  "title": "Complete project documentation",
  "description": "Write API docs for frontend team",
  "priority": "high",
  "due_date": "2025-10-15T10:00:00Z"
}
```
**Response:** `201 Created`
```json
{
  "message": "Todo created successfully",
  "todo": {
    "id": 1,
    "title": "Complete project documentation",
    "priority": "high"
  }
}
```

### Get All Todos
**Endpoint:** `GET /todos?page=1&limit=20&is_completed=false`  
**Authentication:** Required  
**Query Parameters:**
- `page` (optional)
- `limit` (optional)
- `is_completed` (optional)

**Response:** `200 OK`
```json
{
  "todos": [
    {
      "id": 1,
      "title": "Complete project documentation",
      "description": "Write API docs",
      "is_completed": false,
      "priority": "high",
      "due_date": "2025-10-15T10:00:00Z",
      "created_at": "2025-10-11T10:00:00Z"
    }
  ],
  "total": 5,
  "page": 1,
  "limit": 20
}
```

### Get Single Todo
**Endpoint:** `GET /todos/:id`  
**Authentication:** Required  
**Response:** `200 OK`
```json
{
  "id": 1,
  "title": "Complete project documentation",
  "description": "Write API docs",
  "is_completed": false,
  "priority": "high",
  "due_date": "2025-10-15T10:00:00Z"
}
```

### Update Todo
**Endpoint:** `PUT /todos/:id`  
**Authentication:** Required  
**Request Body:**
```json
{
  "title": "Updated title",
  "description": "Updated description",
  "priority": "urgent"
}
```
**Response:** `200 OK`
```json
{
  "message": "Todo updated successfully"
}
```

### Delete Todo
**Endpoint:** `DELETE /todos/:id`  
**Authentication:** Required  
**Response:** `200 OK`
```json
{
  "message": "Todo deleted successfully"
}
```

### Toggle Todo Complete/Incomplete
**Endpoint:** `PATCH /todos/:id/toggle`  
**Authentication:** Required  
**Response:** `200 OK`
```json
{
  "message": "Todo status toggled successfully",
  "is_completed": true
}
```

---

## Admin Routes

> **Note:** All admin routes require admin role

### User Management

#### Get All Users
**Endpoint:** `GET /admin/users?page=1&limit=20`  
**Authentication:** Required (Admin)  
**Response:** `200 OK`
```json
{
  "users": [
    {
      "ID": 1,
      "Name": "John Doe",
      "Email": "john@example.com",
      "PackageID": 2,
      "IS_ACTIVE": true
    }
  ],
  "total": 100,
  "page": 1,
  "limit": 20
}
```

#### Delete User
**Endpoint:** `DELETE /admin/users/:id`  
**Authentication:** Required (Admin)  
**Response:** `200 OK`
```json
{
  "message": "User deleted successfully"
}
```

### Role Management

#### Get All Roles
**Endpoint:** `GET /admin/roles`  
**Authentication:** Required (Admin)  
**Response:** `200 OK`
```json
{
  "roles": [
    {
      "ID": 1,
      "Name": "admin",
      "Permissions": [...]
    },
    {
      "ID": 2,
      "Name": "user",
      "Permissions": [...]
    }
  ]
}
```

#### Get All Permissions
**Endpoint:** `GET /admin/permissions`  
**Authentication:** Required (Admin)  
**Response:** `200 OK`
```json
{
  "permissions": [
    {
      "ID": 1,
      "Name": "CREATE_POST"
    },
    {
      "ID": 2,
      "Name": "DELETE_POST"
    }
  ]
}
```

#### Assign Role to User
**Endpoint:** `POST /admin/users/:id/roles`  
**Authentication:** Required (Admin)  
**Request Body:**
```json
{
  "role_id": 2
}
```
**Response:** `200 OK`
```json
{
  "message": "Role assigned successfully"
}
```

#### Remove Role from User
**Endpoint:** `DELETE /admin/users/:id/roles`  
**Authentication:** Required (Admin)  
**Request Body:**
```json
{
  "role_id": 2
}
```
**Response:** `200 OK`
```json
{
  "message": "Role removed successfully"
}
```

### Blog Management

#### Get All Blogs (Including Unpublished)
**Endpoint:** `GET /admin/blogs?page=1&limit=20`  
**Authentication:** Required (Admin)  
**Response:** `200 OK`
```json
{
  "blogs": [...],
  "total": 50,
  "page": 1,
  "limit": 20
}
```

#### Get Blog Analytics
**Endpoint:** `GET /admin/blogs/analytics`  
**Authentication:** Required (Admin)  
**Response:** `200 OK`
```json
{
  "total_blogs": 50,
  "published_blogs": 45,
  "unpublished_blogs": 5,
  "total_authors": 10,
  "blogs_this_month": 12,
  "blogs_this_week": 3
}
```

### Newsletter Management

#### Get All Newsletter Subscriptions
**Endpoint:** `GET /admin/newsletter/subscriptions?page=1&limit=20`  
**Authentication:** Required (Admin)  
**Response:** `200 OK`
```json
{
  "subscriptions": [
    {
      "id": 1,
      "email": "subscriber@example.com",
      "is_active": true,
      "subscribed_at": "2025-10-01T10:00:00Z"
    }
  ],
  "total": 500,
  "page": 1,
  "limit": 20
}
```

#### Get Newsletter Stats
**Endpoint:** `GET /admin/newsletter/stats`  
**Authentication:** Required (Admin)  
**Response:** `200 OK`
```json
{
  "total_subscriptions": 500,
  "active_subscriptions": 480,
  "inactive_subscriptions": 20,
  "subscriptions_today": 5,
  "subscriptions_this_week": 25,
  "subscriptions_this_month": 100
}
```

### Contact Management

#### Get All Contacts
**Endpoint:** `GET /admin/contacts?page=1&limit=20`  
**Authentication:** Required (Admin)  
**Response:** `200 OK`
```json
{
  "contacts": [
    {
      "id": 1,
      "name": "Jane Doe",
      "email": "jane@example.com",
      "message": "Question about services...",
      "is_read": false,
      "replied": false,
      "created_at": "2025-10-11T10:00:00Z"
    }
  ],
  "total": 50,
  "page": 1,
  "limit": 20
}
```

#### Mark Contact as Read
**Endpoint:** `PUT /admin/contacts/:id/read`  
**Authentication:** Required (Admin)  
**Response:** `200 OK`
```json
{
  "message": "Contact marked as read"
}
```

#### Mark Contact as Replied
**Endpoint:** `PUT /admin/contacts/:id/replied`  
**Authentication:** Required (Admin)  
**Response:** `200 OK`
```json
{
  "message": "Contact marked as replied"
}
```

#### Delete Contact
**Endpoint:** `DELETE /admin/contacts/:id`  
**Authentication:** Required (Admin)  
**Response:** `200 OK`
```json
{
  "message": "Contact deleted successfully"
}
```

#### Get Contact Stats
**Endpoint:** `GET /admin/contacts/stats`  
**Authentication:** Required (Admin)  
**Response:** `200 OK`
```json
{
  "total_contacts": 50,
  "unread_contacts": 10,
  "replied_contacts": 35,
  "contacts_today": 2,
  "contacts