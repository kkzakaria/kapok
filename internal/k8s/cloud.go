package k8s

import (
	"regexp"
	"strings"
)

// CloudProvider represents a supported cloud platform.
type CloudProvider string

const (
	CloudAWS     CloudProvider = "aws"
	CloudGCP     CloudProvider = "gcp"
	CloudAzure   CloudProvider = "azure"
	CloudGeneric CloudProvider = "generic"
)

// CloudConfig holds cloud-specific Kubernetes values.
type CloudConfig struct {
	Provider     CloudProvider
	StorageClass string
	IngressClass string
}

// CloudDetector detects the cloud provider from context.
type CloudDetector interface {
	Detect() CloudProvider
}

// KubeconfigDetector detects cloud provider from a kubeconfig context name.
type KubeconfigDetector struct {
	ContextName string
}

var (
	eksPattern = regexp.MustCompile(`(^|[^a-z])eks([^a-z]|$)`)
	aksPattern = regexp.MustCompile(`(^|[^a-z])aks([^a-z]|$)`)
)

// Detect returns the cloud provider based on the kubeconfig context name.
func (d *KubeconfigDetector) Detect() CloudProvider {
	ctx := strings.ToLower(d.ContextName)

	switch {
	case strings.Contains(ctx, "arn:aws") || eksPattern.MatchString(ctx):
		return CloudAWS
	case strings.Contains(ctx, "gke_") || strings.Contains(ctx, "gke-"):
		return CloudGCP
	case strings.HasSuffix(ctx, ".azmk8s.io") || aksPattern.MatchString(ctx):
		return CloudAzure
	default:
		return CloudGeneric
	}
}

// CloudConfigFor returns cloud-specific configuration values.
func CloudConfigFor(provider CloudProvider) CloudConfig {
	switch provider {
	case CloudAWS:
		return CloudConfig{
			Provider:     CloudAWS,
			StorageClass: "gp3",
			IngressClass: "alb",
		}
	case CloudGCP:
		return CloudConfig{
			Provider:     CloudGCP,
			StorageClass: "standard-rwo",
			IngressClass: "gce",
		}
	case CloudAzure:
		return CloudConfig{
			Provider:     CloudAzure,
			StorageClass: "managed-premium",
			IngressClass: "azure-application-gateway",
		}
	default:
		return CloudConfig{
			Provider:     CloudGeneric,
			StorageClass: "standard",
			IngressClass: "nginx",
		}
	}
}
