package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKubeconfigDetector_Detect(t *testing.T) {
	tests := []struct {
		name     string
		context  string
		expected CloudProvider
	}{
		{"AWS ARN", "arn:aws:eks:us-east-1:123456789:cluster/my-cluster", CloudAWS},
		{"AWS EKS keyword", "my-eks-cluster", CloudAWS},
		{"GCP GKE underscore", "gke_my-project_us-central1_cluster", CloudGCP},
		{"GCP GKE dash", "gke-my-cluster", CloudGCP},
		{"Azure AKS suffix", "my-cluster.azmk8s.io", CloudAzure},
		{"Azure AKS keyword", "my-aks-cluster", CloudAzure},
		{"Generic minikube", "minikube", CloudGeneric},
		{"Generic kind", "kind-kind", CloudGeneric},
		{"Empty context", "", CloudGeneric},
		{"No false positive on desktop (eks)", "my-deksktop-cluster", CloudGeneric},
		{"No false positive on makes (aks)", "makes-cluster", CloudGeneric},
		{"EKS with dashes", "my-eks-cluster", CloudAWS},
		{"AKS with dashes", "my-aks-cluster", CloudAzure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &KubeconfigDetector{ContextName: tt.context}
			assert.Equal(t, tt.expected, d.Detect())
		})
	}
}

func TestCloudConfigFor(t *testing.T) {
	tests := []struct {
		provider     CloudProvider
		storageClass string
		ingressClass string
	}{
		{CloudAWS, "gp3", "alb"},
		{CloudGCP, "standard-rwo", "gce"},
		{CloudAzure, "managed-premium", "azure-application-gateway"},
		{CloudGeneric, "standard", "nginx"},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			cfg := CloudConfigFor(tt.provider)
			assert.Equal(t, tt.provider, cfg.Provider)
			assert.Equal(t, tt.storageClass, cfg.StorageClass)
			assert.Equal(t, tt.ingressClass, cfg.IngressClass)
		})
	}
}
