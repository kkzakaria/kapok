package k8s

import "embed"

// Grafana dashboard JSON files, embedded for maintainability.

//go:embed dashboards/platform-overview.json
var DashboardPlatformOverview string

//go:embed dashboards/per-tenant.json
var DashboardPerTenant string

//go:embed dashboards/graphql.json
var DashboardGraphQL string

//go:embed dashboards/infrastructure.json
var DashboardInfrastructure string

// DashboardFiles provides access to all embedded dashboard files.
//
//go:embed dashboards/*.json
var DashboardFiles embed.FS
