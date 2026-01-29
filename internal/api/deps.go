package api

import (
	"github.com/kapok/kapok/internal/auth"
	"github.com/kapok/kapok/internal/database"
	gql "github.com/kapok/kapok/internal/graphql"
	"github.com/kapok/kapok/internal/tenant"
	"github.com/rs/zerolog"
)

// Dependencies holds all handler dependencies.
type Dependencies struct {
	DB          *database.DB
	JWTManager  *auth.JWTManager
	Provisioner *tenant.Provisioner
	GQLHandler  *gql.Handler
	Logger      zerolog.Logger
}
