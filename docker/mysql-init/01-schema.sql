-- Hey PHP Interpreter - MySQL Test Database Schema
-- This script initializes the test database for MySQLi integration testing

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    age INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock INT DEFAULT 0,
    category VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_category (category),
    INDEX idx_price (price)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL,
    total_price DECIMAL(10, 2) NOT NULL,
    status ENUM('pending', 'processing', 'completed', 'cancelled') DEFAULT 'pending',
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert sample users
INSERT INTO users (name, email, age) VALUES
('Alice Johnson', 'alice@example.com', 28),
('Bob Smith', 'bob@example.com', 35),
('Charlie Brown', 'charlie@example.com', 42),
('Diana Prince', 'diana@example.com', 31),
('Eve Williams', 'eve@example.com', 26);

-- Insert sample products
INSERT INTO products (name, description, price, stock, category) VALUES
('Laptop', 'High-performance laptop with 16GB RAM', 1299.99, 50, 'Electronics'),
('Smartphone', 'Latest model smartphone with 128GB storage', 799.99, 100, 'Electronics'),
('Desk Chair', 'Ergonomic office chair with lumbar support', 249.99, 30, 'Furniture'),
('Coffee Maker', 'Programmable coffee maker with timer', 89.99, 75, 'Appliances'),
('Headphones', 'Noise-cancelling wireless headphones', 199.99, 60, 'Electronics'),
('Desk Lamp', 'LED desk lamp with adjustable brightness', 45.99, 120, 'Furniture'),
('Water Bottle', 'Insulated stainless steel water bottle', 24.99, 200, 'Accessories'),
('Backpack', 'Durable travel backpack with laptop compartment', 79.99, 80, 'Accessories'),
('Mouse', 'Wireless ergonomic mouse', 29.99, 150, 'Electronics'),
('Keyboard', 'Mechanical gaming keyboard with RGB', 129.99, 40, 'Electronics');

-- Insert sample orders
INSERT INTO orders (user_id, product_id, quantity, total_price, status) VALUES
(1, 1, 1, 1299.99, 'completed'),
(1, 5, 1, 199.99, 'completed'),
(2, 2, 2, 1599.98, 'processing'),
(3, 3, 1, 249.99, 'completed'),
(3, 6, 2, 91.98, 'completed'),
(4, 4, 1, 89.99, 'pending'),
(5, 7, 3, 74.97, 'completed'),
(5, 8, 1, 79.99, 'processing'),
(2, 9, 2, 59.98, 'completed'),
(1, 10, 1, 129.99, 'pending');

-- Create a view for user order summary
CREATE OR REPLACE VIEW user_order_summary AS
SELECT
    u.id as user_id,
    u.name as user_name,
    u.email,
    COUNT(o.id) as total_orders,
    SUM(o.total_price) as total_spent,
    MAX(o.order_date) as last_order_date
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
GROUP BY u.id, u.name, u.email;

-- Grant privileges
GRANT ALL PRIVILEGES ON testdb.* TO 'testuser'@'%';
FLUSH PRIVILEGES;
