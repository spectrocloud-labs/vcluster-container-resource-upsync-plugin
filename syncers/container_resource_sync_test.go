package syncers

import (
	"log"
	"testing"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation"
)

var s *containerResourceSyncer

func init() {
	s = &containerResourceSyncer{}
}

func TestUpdateContainerResources(t *testing.T) {
	g := gomega.NewWithT(t)
	tests := []struct {
		pObj *corev1.Pod
		vObj *corev1.Pod
	}{
		{
			pObj: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "longlonglonglonglonglonglonglonglonglonglonglonglonglong",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:              resource.MustParse("1000m"),
									corev1.ResourceMemory:           resource.MustParse("256Mi"),
									corev1.ResourceEphemeralStorage: resource.MustParse("1Gi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:              resource.MustParse("500m"),
									corev1.ResourceMemory:           resource.MustParse("128Mi"),
									corev1.ResourceEphemeralStorage: resource.MustParse("512Mi"),
								},
							},
						},
					},
				},
			},
			vObj: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{},
					},
				},
			},
		},
	}
	for _, test := range tests {
		updated := s.updateContainerResources(test.pObj, test.vObj)
		for k, v := range updated.ObjectMeta.Annotations {
			errs := validation.IsQualifiedName(k)
			g.Expect(errs).To(gomega.BeNil())
			log.Printf("validated annotation: %s: %s", k, v)
		}
	}
}
