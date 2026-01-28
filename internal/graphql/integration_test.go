//go:build integration
// +build integration

package graphql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/kapok/kapok/internal/database"
	"github.com/kapok/kapok/internal/tenant"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq"
)

var (
	testDB       *sql.DB
	testDBConfig database.Config
	pool         *dockertest.Pool
	resource     *dockertest.Resource
)

// setupPostgresContainer creates an ephemeral PostgreSQL container for testing
func setupPostgresContainer(t *testing.T) {
	var err error
	pool, err = dockertest.NewPool("")
	require.NoError(t, err, "could not construct docker pool")

	err = pool.Client.Ping()
	require.NoError(t, err, "could not connect to docker")

	// Pull and run PostgreSQL container
	resource, err = pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=testpass",
			"POSTGRES_USER=testuser",
			"POSTGRES_DB=kapok_graphql_test",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	require.NoError(t, err, "could not start postgres container")

	resource.Expire(120)

	hostAndPort := resource.GetHostPort("5432/tcp")
	dbURL := fmt.Sprintf("postgres://testuser:testpass@%s/kapok_graphql_test?sslmode=disable", hostAndPort)

	err = pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("postgres", dbURL)
		if err != nil {
			return err
		}
		return testDB.Ping()
	})
	require.NoError(t, err, "could not connect to postgres")

	port := resource.GetPort("5432/tcp")
	portInt, err := strconv.Atoi(port)
	require.NoError(t, err)

	testDBConfig = database.Config{
		Host:     "localhost",
		Port:     portInt,
		Database: "kapok_graphql_test",
		User:     "testuser",
		Password: "testpass",
		SSLMode:  "disable",
	}
}

func teardownPostgresContainer(t *testing.T) {
	if testDB != nil {
		testDB.Close()
	}
	if pool != nil && resource != nil {
		pool.Purge(resource)
	}
}

func setupControlDatabase(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS tenants (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(63) UNIQUE NOT NULL,
			schema_name VARCHAR(100) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err)
	
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_log (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id VARCHAR(256),
			action VARCHAR(100) NOT NULL,
			resource VARCHAR(256),
			timestamp TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err)
}

func TestDynamicGraphQLAPI(t *testing.T) {
	setupPostgresContainer(t)
	defer teardownPostgresContainer(t)
	setupControlDatabase(t, testDB)

	ctx := context.Background()
	logger := zerolog.Nop()

	db, err := database.NewDB(ctx, testDBConfig, logger)
	require.NoError(t, err)
	defer db.Close()

	// 1. Create Tenant
	provisioner := tenant.NewProvisioner(db, logger)
	ten, err := provisioner.CreateTenant(ctx, "graphql-test")
	require.NoError(t, err)

	// 2. Create User Table in Tenant Schema manually (simulating migration)
	_, err = testDB.ExecContext(ctx, fmt.Sprintf(`
		CREATE TABLE %s.posts (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT,
			is_published BOOLEAN DEFAULT false
		)
	`, ten.SchemaName))
	require.NoError(t, err)

	// 3. Initialize Handler
	handler := NewHandler(db, logger)

	// 4. Test Mutation (Create Post)
	// Query: mutation { createPosts(title: "Hello World", content: "First post", isPublished: true) { id title } }
	mutationQuery := `
		mutation {
			createPosts(title: "Hello World", content: "First post", isPublished: true) {
				id
				title
				content
				isPublished
			}
		}
	`
	resp := executeGraphQLRequest(t, handler, ten.ID, ten.SchemaName, mutationQuery)
	require.Nil(t, resp.Errors)
	
	var createResult struct {
		CreatePosts struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			Content     string `json:"content"`
			IsPublished bool   `json:"isPublished"`
		} `json:"createPosts"`
	}
	err = json.Unmarshal(resp.Data, &createResult)
	require.NoError(t, err)
	assert.Equal(t, "Hello World", createResult.CreatePosts.Title)
	assert.Equal(t, "First post", createResult.CreatePosts.Content)
	assert.True(t, createResult.CreatePosts.IsPublished)
	assert.Greater(t, createResult.CreatePosts.ID, 0)

	// 5. Test Query (List Posts)
	listQuery := `
		query {
			posts(limit: 10) {
				id
				title
			}
		}
	`
	resp = executeGraphQLRequest(t, handler, ten.ID, ten.SchemaName, listQuery)
	require.Nil(t, resp.Errors)

	var listResult struct {
		Posts []struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
		} `json:"posts"`
	}
	err = json.Unmarshal(resp.Data, &listResult)
	require.NoError(t, err)
	assert.Len(t, listResult.Posts, 1)
	assert.Equal(t, "Hello World", listResult.Posts[0].Title)

	// 6. Test Query Single (Get Post By ID)
	getQuery := fmt.Sprintf(`
		query {
			postsById(id: "%d") {
				title
			}
		}
	`, createResult.CreatePosts.ID)
	resp = executeGraphQLRequest(t, handler, ten.ID, ten.SchemaName, getQuery)
	require.Nil(t, resp.Errors)
	
	var getResult struct {
		PostsById struct {
			Title string `json:"title"`
		} `json:"postsById"`
	}
	err = json.Unmarshal(resp.Data, &getResult)
	require.NoError(t, err)
	assert.Equal(t, "Hello World", getResult.PostsById.Title)

	// 7. Test Mutation (Update Post)
	updateQuery := fmt.Sprintf(`
		mutation {
			updatePosts(id: "%d", title: "Updated Title", content: "Updated content") {
				id
				title
				content
				isPublished
			}
		}
	`, createResult.CreatePosts.ID)
	resp = executeGraphQLRequest(t, handler, ten.ID, ten.SchemaName, updateQuery)
	require.Nil(t, resp.Errors)

	var updateResult struct {
		UpdatePosts struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			Content     string `json:"content"`
			IsPublished bool   `json:"isPublished"`
		} `json:"updatePosts"`
	}
	err = json.Unmarshal(resp.Data, &updateResult)
	require.NoError(t, err)
	assert.Equal(t, createResult.CreatePosts.ID, updateResult.UpdatePosts.ID)
	assert.Equal(t, "Updated Title", updateResult.UpdatePosts.Title)
	assert.Equal(t, "Updated content", updateResult.UpdatePosts.Content)
	assert.True(t, updateResult.UpdatePosts.IsPublished) // Should remain unchanged

	// 8. Test Mutation (Delete Post)
	deleteQuery := fmt.Sprintf(`
		mutation {
			deletePosts(id: "%d") {
				id
				title
			}
		}
	`, createResult.CreatePosts.ID)
	resp = executeGraphQLRequest(t, handler, ten.ID, ten.SchemaName, deleteQuery)
	require.Nil(t, resp.Errors)

	var deleteResult struct {
		DeletePosts struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
		} `json:"deletePosts"`
	}
	err = json.Unmarshal(resp.Data, &deleteResult)
	require.NoError(t, err)
	assert.Equal(t, createResult.CreatePosts.ID, deleteResult.DeletePosts.ID)
	assert.Equal(t, "Updated Title", deleteResult.DeletePosts.Title)

	// 9. Verify Post is deleted
	resp = executeGraphQLRequest(t, handler, ten.ID, ten.SchemaName, listQuery)
	require.Nil(t, resp.Errors)
	err = json.Unmarshal(resp.Data, &listResult)
	require.NoError(t, err)
	assert.Len(t, listResult.Posts, 0)
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

func TestGraphQLRelations(t *testing.T) {
	setupPostgresContainer(t)
	defer teardownPostgresContainer(t)
	setupControlDatabase(t, testDB)

	ctx := context.Background()
	logger := zerolog.Nop()

	db, err := database.NewDB(ctx, testDBConfig, logger)
	require.NoError(t, err)
	defer db.Close()

	// 1. Create Tenant
	provisioner := tenant.NewProvisioner(db, logger)
	ten, err := provisioner.CreateTenant(ctx, "relations-test")
	require.NoError(t, err)

	// 2. Create authors and posts tables with FK relationship
	_, err = testDB.ExecContext(ctx, fmt.Sprintf(`
		CREATE TABLE %s.authors (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT
		)
	`, ten.SchemaName))
	require.NoError(t, err)

	_, err = testDB.ExecContext(ctx, fmt.Sprintf(`
		CREATE TABLE %s.posts (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT,
			author_id INTEGER REFERENCES %s.authors(id)
		)
	`, ten.SchemaName, ten.SchemaName))
	require.NoError(t, err)

	// 3. Initialize Handler
	handler := NewHandler(db, logger)

	// 4. Create an author
	createAuthorQuery := `
		mutation {
			createAuthors(name: "John Doe", email: "john@example.com") {
				id
				name
				email
			}
		}
	`
	resp := executeGraphQLRequest(t, handler, ten.ID, ten.SchemaName, createAuthorQuery)
	require.Nil(t, resp.Errors, "createAuthors errors: %v", resp.Errors)

	var authorResult struct {
		CreateAuthors struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"createAuthors"`
	}
	err = json.Unmarshal(resp.Data, &authorResult)
	require.NoError(t, err)
	assert.Equal(t, "John Doe", authorResult.CreateAuthors.Name)
	authorID := authorResult.CreateAuthors.ID

	// 5. Create a post linked to the author
	createPostQuery := fmt.Sprintf(`
		mutation {
			createPosts(title: "My First Post", content: "Hello World", authorId: %d) {
				id
				title
				authorId
			}
		}
	`, authorID)
	resp = executeGraphQLRequest(t, handler, ten.ID, ten.SchemaName, createPostQuery)
	require.Nil(t, resp.Errors, "createPosts errors: %v", resp.Errors)

	var postResult struct {
		CreatePosts struct {
			ID       int    `json:"id"`
			Title    string `json:"title"`
			AuthorID int    `json:"authorId"`
		} `json:"createPosts"`
	}
	err = json.Unmarshal(resp.Data, &postResult)
	require.NoError(t, err)
	assert.Equal(t, "My First Post", postResult.CreatePosts.Title)
	assert.Equal(t, authorID, postResult.CreatePosts.AuthorID)

	// 6. Test nested query: post with author (belongsTo relation)
	nestedQuery := fmt.Sprintf(`
		query {
			postsById(id: "%d") {
				id
				title
				author {
					id
					name
					email
				}
			}
		}
	`, postResult.CreatePosts.ID)
	resp = executeGraphQLRequest(t, handler, ten.ID, ten.SchemaName, nestedQuery)
	require.Nil(t, resp.Errors, "nested query errors: %v", resp.Errors)

	var nestedResult struct {
		PostsById struct {
			ID     int    `json:"id"`
			Title  string `json:"title"`
			Author struct {
				ID    int    `json:"id"`
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"author"`
		} `json:"postsById"`
	}
	err = json.Unmarshal(resp.Data, &nestedResult)
	require.NoError(t, err)
	assert.Equal(t, "My First Post", nestedResult.PostsById.Title)
	assert.Equal(t, authorID, nestedResult.PostsById.Author.ID)
	assert.Equal(t, "John Doe", nestedResult.PostsById.Author.Name)

	// 7. Test reverse relation: author with posts (hasMany relation)
	hasManyQuery := fmt.Sprintf(`
		query {
			authorsById(id: "%d") {
				id
				name
				posts {
					id
					title
				}
			}
		}
	`, authorID)
	resp = executeGraphQLRequest(t, handler, ten.ID, ten.SchemaName, hasManyQuery)
	require.Nil(t, resp.Errors, "hasMany query errors: %v", resp.Errors)

	var hasManyResult struct {
		AuthorsById struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Posts []struct {
				ID    int    `json:"id"`
				Title string `json:"title"`
			} `json:"posts"`
		} `json:"authorsById"`
	}
	err = json.Unmarshal(resp.Data, &hasManyResult)
	require.NoError(t, err)
	assert.Equal(t, "John Doe", hasManyResult.AuthorsById.Name)
	assert.Len(t, hasManyResult.AuthorsById.Posts, 1)
	assert.Equal(t, "My First Post", hasManyResult.AuthorsById.Posts[0].Title)
}

func executeGraphQLRequest(t *testing.T, handler *Handler, tenantID, schemaName, query string) GraphQLResponse {
	reqBody := fmt.Sprintf(`{"query": %q}`, query)
	req := httptest.NewRequest("POST", "/graphql", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Inject tenant context using the tenant package
	ten := &tenant.Tenant{
		ID:         tenantID,
		SchemaName: schemaName,
		Status:     tenant.StatusActive,
	}
	ctx := tenant.WithTenant(req.Context(), ten)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var gqlResp GraphQLResponse
	err := json.NewDecoder(resp.Body).Decode(&gqlResp)
	require.NoError(t, err)
	return gqlResp
}
