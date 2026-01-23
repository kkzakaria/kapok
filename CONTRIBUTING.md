# Contributing to Kapok

Thank you for your interest in contributing to Kapok!

## Development Workflow

### Branch Protection Rules

**⚠️ CRITICAL: Never commit directly to `main` branch**

All changes MUST go through Pull Requests. Direct commits to `main` are
**blocked** by GitHub branch protection.

### Workflow Steps

1. **Create a feature branch**
   ```bash
   git checkout -b feature/epic-X-story-Y-description
   ```

2. **Make your changes**
   - Follow the coding conventions below
   - Write tests for new features
   - Update documentation as needed

3. **Commit with Conventional Commits format**
   ```bash
   git commit -m "feat(scope): description (Story X.Y)

   - Detailed change 1
   - Detailed change 2

   Epic X, Story Y - dev-complete"
   ```

4. **Push your branch**
   ```bash
   git push -u origin feature/epic-X-story-Y-description
   ```

5. **Create a Pull Request**
   ```bash
   gh pr create --title "feat: Description (Story X.Y)" --body "..."
   ```

6. **Wait for review and merge**
   - PRs are automatically reviewed
   - Once approved, merge to `main`
   - Delete the feature branch after merge

## Commit Message Format

We use **Conventional Commits** for all commit messages.

### Format

```
<type>(<scope>): <description> (Story X.Y)

<body with detailed changes>

Epic X, Story Y - <status>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring without feature changes
- `chore`: Maintenance tasks, dependency updates
- `perf`: Performance improvements
- `style`: Code style changes (formatting, etc.)

### Scope (optional)

- `cli`: CLI commands
- `api`: API changes
- `db`: Database changes
- `k8s`: Kubernetes related
- `docs`: Documentation
- etc.

### Examples

**Feature commit:**

```
feat(cli): Implement Cobra CLI foundation (Story 1.2)

- Added Cobra dependency (v1.10.2)
- Created root command with global flags
- Implemented all command placeholders (init, dev, deploy, tenant)
- Added comprehensive unit tests (all passing)

Epic 1, Story 1.2 - dev-complete
```

**Bug fix:**

```
fix(tenant): Resolve schema isolation leak (Story 2.3)

- Fixed PostgreSQL RLS policy application
- Added integration tests for tenant isolation
- Updated documentation

Epic 2, Story 2.3 - dev-complete
```

**Documentation:**

```
docs: Update README with quick start guide

- Added installation instructions
- Included usage examples
- Fixed broken links
```

## Coding Conventions

### Go Code Style

Follow the patterns defined in
`_bmad-output/planning-artifacts/architecture.md`:

- **Files**: `snake_case.go`
- **Packages**: lowercase, singular, no underscores
- **Structs**: `PascalCase`
- **Functions**: `camelCase` (private), `PascalCase` (exported)
- **Variables**: `camelCase`
- **Constants**: `PascalCase` (exported), `camelCase` (private)

### Logging

- **ALWAYS** use `zerolog` structured logging
- **NEVER** use `fmt.Println()` or `log.Print()`

```go
log.Info().
    Str("tenant_id", tenantID).
    Msg("tenant created successfully")
```

### Error Handling

- Use `%w` for error wrapping to preserve stack traces
- Create custom error types per domain

```go
if err != nil {
    return fmt.Errorf("failed to create tenant: %w", err)
}
```

### Testing

- Co-locate tests with source: `filename_test.go`
- Use table-driven tests
- Use `testify` for assertions

```go
func TestTenantCreate(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid", "test-tenant", false},
        {"empty", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

## Pull Request Guidelines

### PR Title Format

Use the same format as commit messages:

```
feat(scope): Description (Story X.Y)
```

### PR Description Template

```markdown
## Epic X, Story Y: Title

### Changes

- Change 1
- Change 2
- Change 3

### Acceptance Criteria Met

- [x] Criterion 1
- [x] Criterion 2
- [x] Criterion 3

### Testing

- Unit tests: passing
- Integration tests: passing
- Manual testing: completed

### Documentation

- [x] Code comments added
- [x] README updated (if needed)
- [x] Architecture docs updated (if needed)

### Checklist

- [x] Follows coding conventions
- [x] Tests added/updated
- [x] Documentation updated
- [x] No direct commits to main
- [x] Conventional Commits format used
```

### Review Process

1. **Automated Checks**: All PRs must pass:
   - Go tests (`go test ./...`)
   - Linting (`golangci-lint run`)
   - Format check (`gofmt -l .`)

2. **Code Review**: PRs require review (can be auto-approved for simple changes)

3. **Merge**: Use "Squash and merge" to keep main branch history clean

## Architecture & Design Patterns

Before implementing major features:

1. Review `_bmad-output/planning-artifacts/architecture.md`
2. Ensure your design aligns with architectural decisions
3. Follow implementation patterns documented
4. If uncertain, open a discussion issue first

## Questions?

- Check `_bmad-output/planning-artifacts/` for detailed planning docs
- Open an issue for questions
- Tag `@kkzakaria` for urgent matters

## License

_To be determined_
