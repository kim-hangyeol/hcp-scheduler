package resourceinfo

import (
	"Hybrid_Cloud/pkg/apis/hcpcluster/v1alpha1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
)

type ClusterInfo struct {
	ClusterName        string
	Nodes              []*NodeInfo
	RequestedResources *Resources
	AllocableResources *Resources
	CapacityResources  *Resources
	ClusterList        *v1alpha1.HCPClusterList
}

// PodInfo is a wrapper to a Pod with additional pre-computed information to
// accelerate processing. This information is typically immutable (e.g., pre-processed
// inter-pod affinity selectors).
type PodInfo struct {
	ClusterName string
	NodeName    string
	PodName     string

	Pod *v1.Pod

	RequestedResources *Resources
	AllocableResources *Resources
	CapacityResources  *Resources

	RequiredAffinityTerms      []AffinityTerm
	RequiredAntiAffinityTerms  []AffinityTerm
	PreferredAffinityTerms     []WeightedAffinityTerm
	PreferredAntiAffinityTerms []WeightedAffinityTerm
	ParseError                 error
}

// AffinityTerm is a processed version of v1.PodAffinityTerm.
type AffinityTerm struct {
	Namespaces        sets.String
	Selector          labels.Selector
	TopologyKey       string
	NamespaceSelector labels.Selector
}

// WeightedAffinityTerm is a "processed" representation of v1.WeightedAffinityTerm.
type WeightedAffinityTerm struct {
	AffinityTerm
	Weight int32
}

// NodeInfo is node level aggregated information.
type NodeInfo struct {
	ClusterName string
	NodeName    string
	// Overall node information.
	Node *v1.Node

	// Pods running on the node.
	Pods []*PodInfo

	// The subset of pods with affinity.
	PodsWithAffinity []*PodInfo

	// The subset of pods with required anti-affinity.
	PodsWithRequiredAntiAffinity []*PodInfo

	// Total requested resources of all pods on this node.
	RequestedResources   *Resource
	AllocatableResources *Resource
	CapacityResources    *Resource
}

// Resource is a collection of compute resource.
type Resource struct {
	MilliCPU         int64
	Memory           int64
	EphemeralStorage int64
	// We store allowedPodNumber (which is Node.Status.Allocatable.Pods().Value())
	// explicitly as int, to avoid conversions and improve performance.
	AllowedPodNumber int
	// ScalarResources
	ScalarResources map[v1.ResourceName]int64
}

type Resources struct {
	CPU     string
	Memory  string
	Fs      string
	Network string
}