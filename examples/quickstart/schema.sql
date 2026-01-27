-- Sample database schema for Kapok quick start example
-- This creates a simple blog with posts and comments

-- Create posts table
CREATE TABLE IF NOT EXISTS posts (
  id SERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  content TEXT,
  published BOOLEAN DEFAULT false,
  author VARCHAR(100),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Create comments table
CREATE TABLE IF NOT EXISTS comments (
  id SERIAL PRIMARY KEY,
  post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  author VARCHAR(100) NOT NULL,
  content TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Create index on post_id for faster queries
CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);

-- Insert sample data
INSERT INTO posts (title, content, published, author) VALUES
  ('Welcome to Kapok!', 'This is your first blog post created with Kapok. Auto-generated SDKs make it easy to build full-stack applications.', true, 'Kapok Team'),
  ('Getting Started Guide', 'Follow our quick start guide to get up and running in under 5 minutes.', true, 'Kapok Team'),
  ('TypeScript SDK Features', 'The generated TypeScript SDK provides type-safe CRUD operations for all your database tables.', false, 'Developer');

INSERT INTO comments (post_id, author, content) VALUES
  (1, 'Alice', 'Great introduction! Looking forward to using Kapok.'),
  (1, 'Bob', 'The auto-generated SDKs are a game changer!'),
  (2, 'Charlie', 'Followed the guide and had my API running in 3 minutes!');

-- Verify data
SELECT 'Posts created:' as message, COUNT(*) as count FROM posts;
SELECT 'Comments created:' as message, COUNT(*) as count FROM comments;
