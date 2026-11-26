package storageresize

import (
	"context"
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	nsForTest         = "ns1"
	promOpNameForTest = "aPromOpName"
)

func newFakeClientBuilder(t *testing.T) *fake.ClientBuilder {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, appsv1.AddToScheme(scheme))
	require.NoError(t, monitoringv1.AddToScheme(scheme))
	return fake.NewClientBuilder().WithScheme(scheme)
}

func newPVCForTest(t *testing.T, name string, storageSize string, labels map[string]string) corev1.PersistentVolumeClaim {
	storageQuantity, err := resource.ParseQuantity(storageSize)
	require.NoError(t, err)

	return corev1.PersistentVolumeClaim{
		Name:      name,
		Namespace: nsForTest,
		Labels:    labels,
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.VolumeResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: storageQuantity,
				},
			},
		},
	}
}

func newPrometheusForTest(name string, storageSize string, paused bool) *monitoringv1.Prometheus {
	return &monitoringv1.Prometheus{
		Name:      name,
		Namespace: nsForTest,
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
				Paused: paused,
				Storage: &monitoringv1.StorageSpec{
					VolumeClaimTemplate: monitoringv1.EmbeddedPersistentVolumeClaim{
						Spec: corev1.PersistentVolumeClaimSpec{
							Resources: corev1.VolumeResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: resource.MustParse(storageSize),
								},
							},
						},
					},
				},
			},
		},
	}
}

func newResizeStepsForTest(k8sClient client.Client, storageSize string) *defaultResizeSteps {
	return &defaultResizeSteps{
		config: &Config{
			promOperatorName: promOpNameForTest,
			storageSize:      storageSize,
			namespace:        nsForTest,
		},
		k8sClient: k8sClient,
	}
}

type fakeResizeSteps struct {
	configureError         error
	createK8sClientError   error
	patchPVCError          error
	prometheusPauseErrors  []error
	prometheusPauseCount   int
	patchPrometheusError   error
	deleteStatefulSetError error
}

func (f *fakeResizeSteps) configure() error {
	if f.configureError != nil {
		return f.configureError
	}
	return nil
}

func (f *fakeResizeSteps) createK8sClient() error {
	if f.createK8sClientError != nil {
		return f.createK8sClientError
	}
	return nil
}

func (f *fakeResizeSteps) patchPVCs(ctx context.Context) error {
	if f.patchPVCError != nil {
		return f.patchPVCError
	}
	return nil
}

func (f *fakeResizeSteps) setPrometheusPaused(ctx context.Context, pause bool) error {
	if f.prometheusPauseErrors[f.prometheusPauseCount] != nil {
		return f.prometheusPauseErrors[f.prometheusPauseCount]
	}
	f.prometheusPauseCount++
	return nil
}

func (f *fakeResizeSteps) patchPrometheusVolumeClaimTemplate(ctx context.Context) error {
	if f.patchPrometheusError != nil {
		return f.patchPrometheusError
	}
	return nil
}

func (f *fakeResizeSteps) deleteStatefulSet(ctx context.Context) error {
	if f.deleteStatefulSetError != nil {
		return f.deleteStatefulSetError
	}
	return nil
}
