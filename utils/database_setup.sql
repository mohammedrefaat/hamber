-- Database Setup and Initial Data for Hamber Project
-- Run this after your Go application creates the tables

-- Create default roles
INSERT INTO roles (name, created_at, updated_at) VALUES
('admin', NOW(), NOW()),
('moderator', NOW(), NOW()),
('user', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Create default permissions
INSERT INTO permissions (name, created_at, updated_at) VALUES
-- User permissions
('CREATE_BLOG', NOW(), NOW()),
('UPDATE_PROFILE', NOW(), NOW()),
('VIEW_OWN_BLOG', NOW(), NOW()),
('UPLOAD_PHOTOS', NOW(), NOW()),

-- Moderator permissions  
('UPDATE_USER', NOW(), NOW()),
('VIEW_ALL_USERS', NOW(), NOW()),
('MANAGE_BLOG', NOW(), NOW()),
('VIEW_BLOG_ANALYTICS', NOW(), NOW()),

-- Admin permissions
('CREATE_USER', NOW(), NOW()),
('DELETE_USER', NOW(), NOW()),
('ASSIGN_ROLES', NOW(), NOW()),
('MANAGE_NEWSLETTER', NOW(), NOW()),
('MANAGE_CONTACTS', NOW(), NOW()),
('VIEW_ALL_ANALYTICS', NOW(), NOW()),
('MANAGE_PACKAGES', NOW(), NOW()),
('SYSTEM_CONFIG', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Assign permissions to roles
-- Admin gets all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id 
FROM roles r, permissions p 
WHERE r.name = 'admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Moderator permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id 
FROM roles r, permissions p 
WHERE r.name = 'moderator' 
AND p.name IN ('UPDATE_USER', 'VIEW_ALL_USERS', 'MANAGE_BLOG', 'VIEW_BLOG_ANALYTICS', 'CREATE_BLOG', 'UPDATE_PROFILE', 'VIEW_OWN_BLOG', 'UPLOAD_PHOTOS')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- User permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id 
FROM roles r, permissions p 
WHERE r.name = 'user' 
AND p.name IN ('CREATE_BLOG', 'UPDATE_PROFILE', 'VIEW_OWN_BLOG', 'UPLOAD_PHOTOS')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Create default packages
INSERT INTO packages (name, price, duration, benefits, description, is_active, created_at, updated_at, price_per_client) VALUES
('Free Plan', 0.00, 30, '["5 Blog Posts", "Basic Support", "1GB Storage"]', 'Perfect for getting started', true, NOW(), NOW(), false),
('Pro Plan', 29.99, 30, '["Unlimited Blog Posts", "Priority Support", "10GB Storage", "OAuth Integration", "Analytics"]', 'For growing businesses', true, NOW(), NOW(), false),
('Enterprise Plan', 99.99, 30, '["Everything in Pro", "Custom Domains", "Unlimited Storage", "Advanced Analytics", "24/7 Support"]', 'For large organizations', true, NOW(), NOW(), true)
ON CONFLICT DO NOTHING;

-- Create default admin user (password: admin123)
-- Note: In production, change this password immediately!
INSERT INTO users (name, email, password, subdomain, role_id, package_id, is_active, is_email_verified, created_at, updated_at) VALUES
('Administrator', 'admin@hamber.local', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', 1, 2, true, true, NOW(), NOW())
ON CONFLICT (email) DO NOTHING;

-- Assign admin role to admin user
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id 
FROM users u, roles r 
WHERE u.email = 'admin@hamber.local' AND r.name = 'admin'
ON CONFLICT (user_id, role_id) DO NOTHING;

-- Create some sample blog posts
INSERT INTO blogs (title, content, summary, slug, author_id, is_published, published_at, photos, created_at, updated_at) VALUES
('Welcome to Hamber', 'This is your first blog post on Hamber platform. You can create, edit, and manage your blogs easily.', 'Welcome post for new users', 'welcome-to-hamber', 1, true, NOW(), '[]', NOW(), NOW()),
('Getting Started Guide', 'Here's how to get started with Hamber platform...', 'Complete guide for beginners', 'getting-started-guide', 1, true, NOW(), '[]', NOW(), NOW())
ON CONFLICT (slug) DO NOTHING;

-- Verification for setup
SELECT 'Setup completed successfully!' as status;
SELECT COUNT(*) as total_roles FROM roles;
SELECT COUNT(*) as total_permissions FROM permissions; 
SELECT COUNT(*) as total_packages FROM packages;
SELECT COUNT(*) as total_users FROM users;