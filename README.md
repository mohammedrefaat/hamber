# Hamber - Complete Go Web Application

A comprehensive Go web application built with Gin framework, featuring authentication, authorization, blog management, OAuth integration, and more.

## 🚀 Features

- **Authentication & Authorization**
  - JWT-based authentication
  - Role-based access control (RBAC)
  - OAuth integration (Google, Facebook, Apple)
  - Email verification
  - Password reset functionality

- **User Management**
  - User registration and login
  - Profile management
  - Role assignment
  - Permission system

- **Blog System**
  - Create, read, update, delete blogs
  - Image upload with WebP conversion
  - Draft and publish functionality
  - Blog analytics

- **Admin Dashboard**
  - User management
  - Role and permission management
  - Blog management
  - Newsletter management
  - Contact form management
  - Analytics and statistics

- **Additional Features**
  - Newsletter subscription
  - Contact form
  - Rate limiting
  - CORS support
  - Multi-language support

## 🛠️ Installation & Setup

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Git

### 1. Clone Repository

\`\`\`bash
git clone https://github.com/mohammedrefaat/hamber.git
cd hamber
\`\`\`

### 2. Install Dependencies

\`\`\`bash
go mod download
\`\`\`

### 3. Environment Setup

\`\`\`bash
# Copy environment template
cp .env.example .env

# Edit .env file with your configuration
nano .env
\`\`\`

### 4. Database Setup

\`\`\`bash
# Create PostgreSQL database
createdb postgres

# Update config.yaml with your database credentials
# The application will auto-migrate tables on startup

# Run initial data setup
psql -d postgres -f database_setup.sql
\`\`\`

### 5. OAuth Setup (Optional)

To enable OAuth login, set up OAuth apps:

**Google OAuth:**
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Enable Google+ API
4. Create OAuth 2.0 credentials
5. Add redirect URI: `http://localhost:8088/api/auth/oauth/google/callback`
6. Update `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` in `.env`

**Facebook OAuth:**
1. Go to [Facebook Developers](https://developers.facebook.com/)
2. Create a new app
3. Add Facebook Login product
4. Add redirect URI: `http://localhost:8088/api/auth/oauth/facebook/callback`
5. Update `FACEBOOK_CLIENT_ID` and `FACEBOOK_CLIENT_SECRET` in `.env`

**Apple OAuth:**
1. Go to [Apple Developer](https://developer.apple.com/)
2. Create a Service ID
3. Configure Sign in with Apple
4. Add redirect URI: `http://localhost:8088/api/auth/oauth/apple/callback`
5. Update `APPLE_CLIENT_ID` and `APPLE_CLIENT_SECRET` in `.env`

### 6. Run Application

\`\`\`bash
go run main.go
\`\`\`

The server will start on `http://localhost:8088`

## 📚 API Documentation

### Authentication Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/auth/login` | User login | No |
| POST | `/api/auth/register` | User registration | No |
| POST | `/api/auth/refresh` | Refresh JWT token | No |
| POST | `/api/auth/forgot-password` | Request password reset | No |
| POST | `/api/auth/reset-password` | Reset password | No |

### OAuth Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/auth/oauth/google` | Google OAuth login |
| GET | `/api/auth/oauth/google/callback` | Google OAuth callback |
| GET | `/api/auth/oauth/facebook` | Facebook OAuth login |
| GET | `/api/auth/oauth/facebook/callback` | Facebook OAuth callback |
| GET | `/api/auth/oauth/apple` | Apple OAuth login |
| GET | `/api/auth/oauth/apple/callback` | Apple OAuth callback |

### User & Profile Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/profile` | Get user profile | Yes |
| PUT | `/api/profile` | Update user profile | Yes |
| **GET** | **`/api/permissions`** | **Get user permissions** | **Yes (Token Only)** |

### Admin Endpoints

| Method | Endpoint | Description | Auth Required | Role Required |
|--------|----------|-------------|---------------|---------------|
| GET | `/api/admin/users` | Get all users | Yes | Admin |
| DELETE | `/api/admin/users/:id` | Delete user | Yes | Admin |
| POST | `/api/admin/users/:id/roles` | Assign role to user | Yes | Admin |
| DELETE | `/api/admin/users/:id/roles` | Remove role from user | Yes | Admin |
| GET | `/api/admin/roles` | Get all roles | Yes | Admin |
| GET | `/api/admin/permissions` | Get all permissions | Yes | Admin |

### Blog Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/blogs` | Get public blogs | No |
| GET | `/api/blogs/:id` | Get specific blog | No |
| POST | `/api/blogs` | Create blog | Yes |
| PUT | `/api/blogs/:id` | Update blog | Yes |
| DELETE | `/api/blogs/:id` | Delete blog | Yes |
| POST | `/api/blogs/:id/photos` | Upload blog photos | Yes |

### Package Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/packages` | Get all packages | No |
| GET | `/api/packages/:id` | Get specific package | No |

### Newsletter & Contact Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/newsletter/subscribe` | Subscribe to newsletter | No |
| POST | `/api/newsletter/unsubscribe` | Unsubscribe from newsletter | No |
| POST | `/api/contact` | Submit contact form | No |

## 🔐 Permission System

The application uses a comprehensive role-based permission system:

### Default Roles

- **Admin**: Full system access
- **Moderator**: User and blog management
- **User**: Basic user functionality

### Permission Endpoint Usage

The new permission endpoint (`GET /api/permissions`) allows you to get user permissions using only the JWT token in the Authorization header:

\`\`\`bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" http://localhost:8088/api/permissions
\`\`\`

Response:
\`\`\`json
{
  "user_permissions": {
    "user_id": 1,
    "email": "user@example.com", 
    "role": "admin",
    "permissions": [
      {"id": 1, "name": "CREATE_USER"},
      {"id": 2, "name": "DELETE_USER"},
      // ... more permissions
    ]
  },
  "jwt_permissions": ["CREATE_USER", "DELETE_USER", "..."],
  "message": "Permissions retrieved successfully"
}
\`\`\`

## 🔧 Configuration

### Environment Variables

Key environment variables:

- `JWT_SECRET`: Secret key for JWT tokens
- `JWT_EXPIRATION_HOURS`: JWT expiration time
- `GOOGLE_CLIENT_ID`: Google OAuth client ID
- `FACEBOOK_CLIENT_ID`: Facebook OAuth app ID
- `APPLE_CLIENT_ID`: Apple OAuth service ID

### Database Configuration

Update `config.yaml`:

\`\`\`yaml
database:
  host: localhost
  user: postgres
  password: yourpassword
  dbname: hamber
  port: 5432
\`\`\`

## 🚦 Testing

### Default Admin Account

- **Email**: `admin@hamber.local`
- **Password**: `admin123`

**⚠️ Important**: Change the admin password immediately in production!

### API Testing

Use the provided Postman collection or test with curl:

\`\`\`bash
# Login
curl -X POST http://localhost:8088/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@hamber.local","password":"admin123"}'

# Get permissions (replace TOKEN with actual JWT)
curl -H "Authorization: Bearer TOKEN" http://localhost:8088/api/permissions
\`\`\`

## 📁 Project Structure

\`\`\`
hamber/
├── Config/           # Configuration management
├── controllers/      # HTTP request handlers
├── DB_models/        # Database models
├── Db/              # Database connection
├── Middleware/       # HTTP middleware
├── services/         # Business logic and routing
├── stores/           # Data access layer
├── Tools/            # Utility functions
├── utils/            # Helper utilities
├── version/          # Version information
├── config.yaml       # Application configuration
├── main.go           # Application entry point
└── README.md         # This file
\`\`\`

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

If you encounter any issues:

1. Check the logs for error messages
2. Ensure all environment variables are set correctly
3. Verify database connection and migrations
4. Check OAuth configuration if using OAuth features

## ✨ What's Fixed & New

### ✅ Fixed Issues

- **OAuth Configuration**: Added missing `InitOAuthConfig()` function
- **JWT Role Field**: Fixed commented-out role field in JWT generation
- **GetAllUsers**: Implemented missing method in stores and controller
- **Permission System**: Complete role-based permission system
- **Apple OAuth**: Fixed configuration and implementation

### 🆕 New Features

- **Permission Endpoint**: `GET /api/permissions` for token-only permission retrieval
- **Role Management**: Admin endpoints for role assignment
- **Enhanced Security**: Comprehensive permission validation
- **Better Error Handling**: Improved error responses
- **Database Setup**: Automated initial data setup script

The application is now fully functional with all errors fixed and requested features implemented!