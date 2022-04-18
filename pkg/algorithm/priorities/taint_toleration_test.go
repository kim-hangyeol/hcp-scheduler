package priorities

import (
	"Hybrid_Cloud/hcp-scheduler/pkg/resourceinfo"
	"Hybrid_Cloud/pkg/apis/resource/v1alpha1"
	"fmt"
	"strconv"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func nodeWithTaints(nodeName string, taints []v1.Taint) *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
		},
		Spec: v1.NodeSpec{
			Taints: taints,
		},
	}
}

func podWithTolerations(tolerations []v1.Toleration) *v1.Pod {
	return &v1.Pod{
		Spec: v1.PodSpec{
			Tolerations: tolerations,
		},
	}
}

// This function will create a set of nodes and pods and test the priority
// Nodes with zero,one,two,three,four and hundred taints are created
// Pods with zero,one,two,three,four and hundred tolerations are created

func TestTaintAndToleration(t *testing.T) {
	tests := []struct {
		pod  *v1.Pod
		node []*v1.Node
	}{
		{
			podWithTolerations([]v1.Toleration{{
				Key:      "foo",
				Operator: v1.TolerationOpEqual,
				Value:    "bar",
				Effect:   v1.TaintEffectPreferNoSchedule,
			}}),
			[]*v1.Node{
				nodeWithTaints("nodeA", []v1.Taint{{
					Key:    "foo",
					Value:  "bar",
					Effect: v1.TaintEffectPreferNoSchedule,
				}}),
				nodeWithTaints("nodeB", []v1.Taint{{
					Key:    "foo",
					Value:  "blah",
					Effect: v1.TaintEffectPreferNoSchedule,
				}}),
			},
		},

		{ // the count of taints that are tolerated by pod, does not matter.
			podWithTolerations([]v1.Toleration{
				{
					Key:      "cpu-type",
					Operator: v1.TolerationOpEqual,
					Value:    "arm64",
					Effect:   v1.TaintEffectPreferNoSchedule,
				}, {
					Key:      "disk-type",
					Operator: v1.TolerationOpEqual,
					Value:    "ssd",
					Effect:   v1.TaintEffectPreferNoSchedule,
				},
			}),
			[]*v1.Node{
				nodeWithTaints("nodeA", []v1.Taint{}),
				nodeWithTaints("nodeB", []v1.Taint{
					{
						Key:    "cpu-type",
						Value:  "arm64",
						Effect: v1.TaintEffectPreferNoSchedule,
					},
				}),
				nodeWithTaints("nodeC", []v1.Taint{
					{
						Key:    "cpu-type",
						Value:  "arm64",
						Effect: v1.TaintEffectPreferNoSchedule,
					}, {
						Key:    "disk-type",
						Value:  "ssd",
						Effect: v1.TaintEffectPreferNoSchedule,
					},
				}),
			},
		},
		{ // the count of taints on a node that are not tolerated by pod, matters.
			podWithTolerations([]v1.Toleration{{
				Key:      "foo",
				Operator: v1.TolerationOpEqual,
				Value:    "bar",
				Effect:   v1.TaintEffectPreferNoSchedule,
			}}),
			[]*v1.Node{
				nodeWithTaints("nodeA", []v1.Taint{}),
				nodeWithTaints("nodeB", []v1.Taint{
					{
						Key:    "cpu-type",
						Value:  "arm64",
						Effect: v1.TaintEffectPreferNoSchedule,
					},
				}),
				nodeWithTaints("nodeC", []v1.Taint{
					{
						Key:    "cpu-type",
						Value:  "arm64",
						Effect: v1.TaintEffectPreferNoSchedule,
					}, {
						Key:    "disk-type",
						Value:  "ssd",
						Effect: v1.TaintEffectPreferNoSchedule,
					},
				}),
			},
		},
		{ // taints-tolerations priority only takes care about the taints and tolerations that have effect PreferNoSchedule
			podWithTolerations([]v1.Toleration{
				{
					Key:      "cpu-type",
					Operator: v1.TolerationOpEqual,
					Value:    "arm64",
					Effect:   v1.TaintEffectNoSchedule,
				}, {
					Key:      "disk-type",
					Operator: v1.TolerationOpEqual,
					Value:    "ssd",
					Effect:   v1.TaintEffectNoSchedule,
				},
			}),
			[]*v1.Node{
				nodeWithTaints("nodeA", []v1.Taint{}),
				nodeWithTaints("nodeB", []v1.Taint{
					{
						Key:    "cpu-type",
						Value:  "arm64",
						Effect: v1.TaintEffectNoSchedule,
					},
				}),
				nodeWithTaints("nodeC", []v1.Taint{
					{
						Key:    "cpu-type",
						Value:  "arm64",
						Effect: v1.TaintEffectPreferNoSchedule,
					}, {
						Key:    "disk-type",
						Value:  "ssd",
						Effect: v1.TaintEffectPreferNoSchedule,
					},
				}),
			},
		},
		{
			podWithTolerations([]v1.Toleration{}),
			[]*v1.Node{
				//Node without taints
				nodeWithTaints("nodeA", []v1.Taint{}),
				nodeWithTaints("nodeB", []v1.Taint{
					{
						Key:    "cpu-type",
						Value:  "arm64",
						Effect: v1.TaintEffectPreferNoSchedule,
					},
				}),
			},
		},
	}

	var clusterinfo_list resourceinfo.ClusterInfoList
	for i, test := range tests {
		nodes_list := test.node
		cluster_name := "test_cluster" + strconv.Itoa(i+1)
		createTestClusters(&clusterinfo_list, nodes_list, cluster_name)
	}

	var rep int32 = 2

	test := tests[1].pod
	fmt.Println("TaintToleration")
	test_deployment := &v1alpha1.HCPDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test_deployment",
			Annotations: map[string]string{},
		},
		Spec: v1alpha1.HCPDeploymentSpec{
			RealDeploymentSpec: appsv1.DeploymentSpec{
				Replicas: &rep,
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Tolerations: test.Spec.Tolerations,
					},
				},
			},
		},
	}

	replicas := *test_deployment.Spec.RealDeploymentSpec.Replicas

	fake_pod := newPodFromHCPDeployment(test_deployment)

	for i := 0; i < int(replicas); i++ {
		scoring(fake_pod, &clusterinfo_list, "TaintToleration")
	}

}