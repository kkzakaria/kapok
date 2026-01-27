# Quick Start Example

This directory contains a complete working example to help you get started with
Kapok.

## What's Included

- `schema.sql` - Sample database schema (blog with posts and comments)
- `config.yaml` - Sample configuration file
- `client-example.ts` - TypeScript SDK usage examples
- `react-example.tsx` - React hooks usage examples

## How to Use

### 1. Create Database

```bash
# Create database
createdb my_blog_example

# Run schema
psql -d my_blog_example -f schema.sql
```

### 2. Configure Kapok

Copy `config.yaml` to your project or set environment variables:

```bash
export KAPOK_DATABASE_DATABASE="my_blog_example"
export KAPOK_DATABASE_PASSWORD="your_password"
```

### 3. Start Kapok

```bash
kapok dev
```

### 4. Generate SDKs

```bash
# TypeScript SDK
kapok generate sdk --schema public --project-name blog-sdk

# React hooks
kapok generate react --sdk-import ../typescript
```

### 5. Try the Examples

#### TypeScript Example

```bash
cd sdk/typescript
npm install
npm run build
cd ../..

# Copy and run the example
cp examples/quickstart/client-example.ts .
npx ts-node client-example.ts
```

#### React Example

```bash
# Create a new React app
npx create-react-app my-blog-frontend
cd my-blog-frontend

# Install SDKs
npm install ../sdk/typescript ../sdk/react @tanstack/react-query

# Copy example component
cp ../examples/quickstart/react-example.tsx src/App.tsx

# Start app
npm start
```

## Expected Output

### TypeScript Example

```
Created post: 1
Found 1 posts
Post #1: Hello from TypeScript SDK!
Updated post: Hello from TypeScript SDK! (Updated)
Post deleted successfully
```

### React Example

Opens a browser showing:

- List of all blog posts
- Ability to create new posts
- Real-time updates when posts are added

## Troubleshooting

### Database Connection Error

Make sure PostgreSQL is running and the database exists:

```bash
psql -l | grep my_blog_example
```

### SDK Not Found

Ensure you've generated and built the SDKs:

```bash
ls -la sdk/typescript/dist
ls -la sdk/react/dist
```

## Next Steps

- Explore the generated SDK code in `sdk/typescript/src/`
- Check out the React hooks in `sdk/react/src/hooks/`
- Modify the schema and regenerate SDKs
- Build your own frontend application!

---

**Questions?** See the [Quick Start Guide](../../docs/quickstart.md) or
[open an issue](https://github.com/kapok/kapok/issues).
