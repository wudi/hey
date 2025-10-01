-- MySQL initialization script for PDO tests

-- Create test tables
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL,
    age INT,
    balance DECIMAL(10, 2) DEFAULT 0.00,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS posts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    published BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS tags (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Insert sample data
INSERT INTO users (username, email, age, balance, is_active) VALUES
('john_doe', 'john@example.com', 30, 1000.50, TRUE),
('jane_smith', 'jane@example.com', 25, 2500.75, TRUE),
('bob_wilson', 'bob@example.com', 35, 500.00, FALSE),
('alice_brown', 'alice@example.com', 28, 3000.00, TRUE),
('charlie_davis', 'charlie@example.com', 42, 750.25, TRUE);

INSERT INTO posts (user_id, title, content, published) VALUES
(1, 'First Post', 'This is my first blog post', TRUE),
(1, 'Second Post', 'Another day, another post', TRUE),
(2, 'Hello World', 'Just getting started', TRUE),
(2, 'Draft Post', 'This is not published yet', FALSE),
(4, 'Alice Adventures', 'My journey begins here', TRUE);

INSERT INTO tags (name) VALUES
('php'),
('mysql'),
('programming'),
('web-development'),
('tutorial');

-- Grant privileges
GRANT ALL PRIVILEGES ON testdb.* TO 'testuser'@'%';
FLUSH PRIVILEGES;
