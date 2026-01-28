package graphql

import (
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/iancoleman/strcase"
)

// SchemaGenerator generates a GraphQL schema from database metadata
type SchemaGenerator struct {
	resolver *Resolver
}

// NewSchemaGenerator creates a new schema generator
func NewSchemaGenerator(resolver *Resolver) *SchemaGenerator {
	return &SchemaGenerator{resolver: resolver}
}

// Generate creates a GraphQL schema for a specific tenant schema
func (g *SchemaGenerator) Generate(tenantSchema string, metadata *SchemaMetadata) (*graphql.Schema, error) {
	// 1. Create GraphQL Objects for each table
	types := make(map[string]*graphql.Object)

	// Build a map of table name -> Table for quick lookup
	tableMap := make(map[string]Table)
	for _, table := range metadata.Tables {
		tableMap[table.Name] = table
	}

	for _, table := range metadata.Tables {
		table := table // capture loop variable
		typeName := strcase.ToCamel(table.Name)

		types[table.Name] = graphql.NewObject(graphql.ObjectConfig{
			Name: typeName,
			Fields: (graphql.FieldsThunk)(func() graphql.Fields {
				fields := graphql.Fields{}

				// Add columns as fields
				for _, col := range table.Columns {
					fieldName := strcase.ToLowerCamel(col.Name)
					gqlType := g.getGraphQLType(col.DataType)
					if !col.IsNullable {
						gqlType = graphql.NewNonNull(gqlType)
					}

					fields[fieldName] = &graphql.Field{
						Type: gqlType,
					}

					// Add FK relation field (e.g., author for author_id)
					if col.IsFK && col.FKTable != "" {
						relatedType, exists := types[col.FKTable]
						if exists {
							// Remove _id suffix for relation field name
							relationFieldName := strings.TrimSuffix(fieldName, "Id")
							if relationFieldName == fieldName {
								// No Id suffix, use table name
								relationFieldName = strcase.ToLowerCamel(col.FKTable)
							}

							fields[relationFieldName] = &graphql.Field{
								Type:    relatedType,
								Resolve: g.resolver.ResolveRelation(tenantSchema, col.FKTable, col.FKColumn, col.Name),
							}
						}
					}
				}

				// Add reverse relations (hasMany) - e.g., posts for a user
				for otherTableName, otherTable := range tableMap {
					if otherTableName == table.Name {
						continue
					}
					for _, otherCol := range otherTable.Columns {
						if otherCol.IsFK && otherCol.FKTable == table.Name {
							// This table has a FK pointing to current table
							// Add a "hasMany" relation field
							relatedType, exists := types[otherTableName]
							if exists {
								// Field name is plural of the related table
								hasManyFieldName := strcase.ToLowerCamel(otherTableName)

								fields[hasManyFieldName] = &graphql.Field{
									Type: graphql.NewList(relatedType),
									Args: graphql.FieldConfigArgument{
										"limit": &graphql.ArgumentConfig{
											Type: graphql.Int,
										},
										"offset": &graphql.ArgumentConfig{
											Type: graphql.Int,
										},
									},
									Resolve: g.resolver.ResolveHasMany(tenantSchema, otherTableName, otherCol.Name, otherCol.FKColumn),
								}
							}
						}
					}
				}

				return fields
			}),
		})
	}

	// 2. Create Query Root
	queryFields := graphql.Fields{}
	for _, table := range metadata.Tables {
		tableName := table.Name
		fieldName := strcase.ToLowerCamel(tableName)
		gqlType, ok := types[tableName]
		if !ok {
			continue
		}
		pkName := g.getPrimaryKey(table)

		// List Query: users(limit: Int, offset: Int)
		queryFields[fieldName] = &graphql.Field{
			Type: graphql.NewList(gqlType),
			Args: graphql.FieldConfigArgument{
				"limit": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
				"offset": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
			},
			Resolve: g.resolver.ResolveList(tenantSchema, tableName),
		}

		// Get Query: userById(id: ID!)
		if pkName != "" {
			singleFieldName := fieldName + "ById"
			queryFields[singleFieldName] = &graphql.Field{
				Type: gqlType,
				Args: graphql.FieldConfigArgument{
					pkName: &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: g.resolver.ResolveGet(tenantSchema, tableName, pkName),
			}
		}
	}

	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Query",
		Fields: queryFields,
	})

	// 3. Create Mutation Root
	mutationFields := graphql.Fields{}
	for _, table := range metadata.Tables {
		tableName := table.Name
		typeName := strcase.ToCamel(tableName)
		gqlType := types[tableName]
		pkName := g.getPrimaryKey(table)

		// Create Mutation: createBroad(name: String!, ...)
		createArgs := graphql.FieldConfigArgument{}
		var createCols []string

		for _, col := range table.Columns {
			argName := strcase.ToLowerCamel(col.Name)
			argType := g.getGraphQLType(col.DataType)
			
			if !col.IsNullable && col.Name != pkName && col.Name != "created_at" && col.Name != "updated_at" {
				argType = graphql.NewNonNull(argType)
			}
			
			createArgs[argName] = &graphql.ArgumentConfig{
				Type: argType,
			}
			createCols = append(createCols, col.Name)
		}

		mutationFields["create"+typeName] = &graphql.Field{
			Type:    gqlType,
			Args:    createArgs,
			Resolve: g.resolver.ResolveCreate(tenantSchema, tableName, createCols),
		}

		// Update Mutation: updatePosts(id: ID!, title: String, ...)
		if pkName != "" {
			updateArgs := graphql.FieldConfigArgument{
				pkName: &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.ID),
				},
			}
			var updateCols []string

			for _, col := range table.Columns {
				argName := strcase.ToLowerCamel(col.Name)
				// Skip PK in update args (already added as required)
				if col.Name == pkName {
					updateCols = append(updateCols, col.Name)
					continue
				}
				argType := g.getGraphQLType(col.DataType)
				// All fields are optional for updates
				updateArgs[argName] = &graphql.ArgumentConfig{
					Type: argType,
				}
				updateCols = append(updateCols, col.Name)
			}

			mutationFields["update"+typeName] = &graphql.Field{
				Type:    gqlType,
				Args:    updateArgs,
				Resolve: g.resolver.ResolveUpdate(tenantSchema, tableName, pkName, updateCols),
			}

			// Delete Mutation: deletePosts(id: ID!)
			mutationFields["delete"+typeName] = &graphql.Field{
				Type: gqlType,
				Args: graphql.FieldConfigArgument{
					pkName: &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: g.resolver.ResolveDelete(tenantSchema, tableName, pkName),
			}
		}
	}

	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Mutation",
		Fields: mutationFields,
	})

	schemaConfig := graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	}

	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

func (g *SchemaGenerator) getGraphQLType(dataType string) graphql.Type {
	dataType = strings.ToLower(dataType)
	switch {
	case strings.Contains(dataType, "int"):
		return graphql.Int
	case strings.Contains(dataType, "char") || strings.Contains(dataType, "text") || strings.Contains(dataType, "uuid"):
		return graphql.String
	case strings.Contains(dataType, "bool"):
		return graphql.Boolean
	case strings.Contains(dataType, "float") || strings.Contains(dataType, "double") || strings.Contains(dataType, "numeric") || strings.Contains(dataType, "decimal"):
		return graphql.Float
	case strings.Contains(dataType, "time") || strings.Contains(dataType, "date"):
		return graphql.String
	default:
		return graphql.String
	}
}

func (g *SchemaGenerator) getPrimaryKey(table Table) string {
	for _, col := range table.Columns {
		if col.IsPK {
			return col.Name
		}
	}
	// Fallback to "id" if found
	for _, col := range table.Columns {
		if col.Name == "id" {
			return col.Name
		}
	}
	return ""
}
