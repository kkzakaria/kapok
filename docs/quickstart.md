# Kapok Quick Start Guide

Get a fully functional backend with auto-generated TypeScript SDK and React
hooks in under 5 minutes.

## Prerequisites

Before starting, ensure you have:

- **Go 1.21+** - [Download here](https://golang.org/dl/)
- **PostgreSQL** running locally or remotely
- **Node.js 18+** (for SDK generation) - [Download here](https://nodejs.org/)

Verify your setup:

```bash
go version        # Should show 1.21 or higher
psql --version    # Verify PostgreSQL is installed
node --version    # Should show 18.0 or higher
```

## Installation

### Option 1: Install from Source (Recommended)

```bash
go install github.com/kapok/kapok/cmd/kapok@latest
```

Verify installation:

```bash
kapok version
```

### Option 2: Download Binary

See [Installation Guide](./installation.md) for platform-specific binaries.

## Step 1: Initialize Your Project (1 minute)

Create a new Kapok project:

```bash
kapok init my-blog
cd my-blog
```

This creates:

```
my-blog/
├── config.yaml           # Configuration file
├── migrations/           # Database migrations
└── ...
```

**Configure database** in `config.yaml`:

```yaml
database:
    host: localhost
    port: 5432
    user: postgres
    password: your_password
    database: my_blog
    ssl_mode: disable
```

Or use environment variables:

```bash
export KAPOK_DATABASE_PASSWORD="your_password"
export KAPOK_DATABASE_DATABASE="my_blog"
```

## Step 2: Start Development Server (1 minute)

Start the Kapok development environment:

```bash
kapok dev
```

This will:

- ✅ Connect to PostgreSQL
- ✅ Run database migrations
- ✅ Start the API server on `http://localhost:8080`

You should see:

```
INFO Starting Kapok development server...
INFO Connected to PostgreSQL
INFO Running migrations...
INFO API server listening on :8080
```

Keep this terminal running and open a new one for the next steps.

## Step 3: Create Your Schema (1 minute)

Create a sample database schema. In your database, run:

```sql
CREATE TABLE posts (
  id SERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  content TEXT,
  published BOOLEAN DEFAULT false,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE comments (
  id SERIAL PRIMARY KEY,
  post_id INTEGER REFERENCES posts(id),
  author VARCHAR(100),
  content TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);
```

Or use the example from `examples/quickstart/schema.sql`.

## Step 4: Generate TypeScript SDK (30 seconds)

Generate a type-safe TypeScript client:

```bash
kapok generate sdk --schema public --project-name my-blog-sdk
```

This creates `sdk/typescript/` with:

- TypeScript interfaces for all tables
- CRUD functions (`createPosts`, `listPosts`, etc.)
- Type-safe client class

**Install and build**:

```bash
cd sdk/typescript
npm install
npm run build
cd ../..
```

## Step 5: Generate React Hooks (30 seconds)

Generate React Query hooks:

```bash
kapok generate react --sdk-import ../typescript
```

This creates `sdk/react/` with hooks like:

- `useListPosts()` - Query posts with caching
- `useCreatePosts()` - Create posts with auto-refresh
- `usePostsById(id)` - Fetch single post

**Install and build**:

```bash
cd sdk/react
npm install
npm run build
cd ../..
```

## Step 6: Use Your API (1 minute)

### Option A: Direct REST API

```bash
# Create a post
curl -X POST http://localhost:8080/api/posts \
  -H "Content-Type: application/json" \
  -d '{"title":"Hello Kapok!","content":"My first post","published":true}'

# List all posts
curl http://localhost:8080/api/posts
```

### Option B: TypeScript SDK

```typescript
import { KapokClient } from "my-blog-sdk";

const client = new KapokClient("http://localhost:8080/api");

// Create a post
const post = await client.posts.create({
    title: "Hello Kapok!",
    content: "My first blog post",
    published: true,
});

console.log("Created post:", post.id);

// List all posts
const allPosts = await client.posts.list({ limit: 10 });
console.log(`Found ${allPosts.length} posts`);

// Get single post
const singlePost = await client.posts.getById(1);
```

### Option C: React Hooks

```typescript
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { KapokProvider, useCreatePosts, useListPosts } from "my-blog-sdk-react";

const queryClient = new QueryClient();

function App() {
    return (
        <QueryClientProvider client={queryClient}>
            <KapokProvider baseUrl="http://localhost:8080/api">
                <BlogPosts />
            </KapokProvider>
        </QueryClientProvider>
    );
}

function BlogPosts() {
    const { data: posts, isLoading, error } = useListPosts();
    const createPost = useCreatePosts();

    if (isLoading) return <div>Loading...</div>;
    if (error) return <div>Error: {error.message}</div>;

    return (
        <div>
            <h1>My Blog Posts</h1>
            {posts?.map((post) => (
                <article key={post.id}>
                    <h2>{post.title}</h2>
                    <p>{post.content}</p>
                    <small>
                        Created: {new Date(post.createdAt).toLocaleDateString()}
                    </small>
                </article>
            ))}

            <button
                onClick={() =>
                    createPost.mutate({
                        title: "New Post",
                        content: "Created with React hooks!",
                        published: true,
                    })}
            >
                Add Post
            </button>
        </div>
    );
}
```

## 🎉 Congratulations!

You now have:

- ✅ A running Kapok backend
- ✅ Auto-generated TypeScript SDK
- ✅ React hooks with caching
- ✅ Type-safe API client

## What's Next?

- **Add Authentication**: Implement JWT auth
- **Add More Tables**: Extend your schema
- **Deploy**: Deploy to production
- **Customize**: Configure middleware, logging, etc.

## Troubleshooting

### Database Connection Failed

**Problem**: `failed to connect to database`

**Solution**:

1. Verify PostgreSQL is running: `psql -U postgres`
2. Check `config.yaml` credentials
3. Try setting environment variables:
   ```bash
   export KAPOK_DATABASE_PASSWORD="your_password"
   export KAPOK_DATABASE_DATABASE="my_blog"
   ```

### Port Already in Use

**Problem**: `address already in use: :8080`

**Solution**:

1. Change port in `config.yaml`:
   ```yaml
   server:
       port: 8081
   ```
2. Or kill the process using port 8080:
   ```bash
   lsof -ti:8080 | xargs kill -9
   ```

### SDK Generation Failed

**Problem**: `no tables found in schema`

**Solution**:

1. Verify tables exist: `psql -U postgres -d my_blog -c "\dt"`
2. Check schema name (default is `public`)
3. Ensure you're connected to the correct database

### TypeScript Compilation Errors

**Problem**: Generated SDK has type errors

**Solution**:

1. Ensure TypeScript 5+ is installed: `npm install -D typescript@^5.0.0`
2. Regenerate the SDK: `kapok generate sdk`
3. Clean build: `rm -rf dist && npm run build`

### React Hooks Not Working

**Problem**: `useKapokClient must be used within a KapokProvider`

**Solution**: Wrap your app with providers:

```typescript
<QueryClientProvider client={queryClient}>
    <KapokProvider baseUrl="http://localhost:8080/api">
        {/* Your app here */}
    </KapokProvider>
</QueryClientProvider>;
```

## Common Issues

### "command not found: kapok"

Add `$GOPATH/bin` to your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

Add to `~/.bashrc` or `~/.zshrc` to make permanent.

### Windows WSL Issues

If using Windows Subsystem for Linux, ensure:

1. PostgreSQL is running in WSL, not Windows
2. Use `localhost` not `127.0.0.1`
3. Firewall allows connections

## Need Help?

- 📖 [Full Documentation](../README.md)
- 🐛 [Report Issues](https://github.com/kapok/kapok/issues)
- 💬 [Discussions](https://github.com/kapok/kapok/discussions)

---

**Time to complete**: ~5 minutes ⏱️\
**Difficulty**: Beginner 🟢
