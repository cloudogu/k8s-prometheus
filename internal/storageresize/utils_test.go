package storageresize

import (
	"context"
	"reflect"
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

const (
	nsForTest         = "ns1"
	promOpNameForTest = "aPromOpName"
)

type promPaused int

const (
	Unknown promPaused = iota
	Paused
	NotPaused
)

type callCounter struct {
	listCount                    int
	listError                    error
	getCount                     int
	getError                     error
	patchCount                   int
	patchError                   error
	patchedPVCs                  []string
	patchedPVCStorageQuantities  []*resource.Quantity
	patchedPromPaused            promPaused
	patchedPromStorageQuantity   *resource.Quantity
	deleteAllOfError             error
	deleteAllOfType              string
	deleteAllOfTypenNamespace    string
	deleteAllOfPropagationPolicy metav1.DeletionPropagation
	deleteAllOfLabelSelector     string
}

func (c *callCounter) getCall(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if c.getError != nil {
		return c.getError
	}
	c.getCount++
	return client.Get(ctx, key, obj, opts...)
}

func (c *callCounter) listCall(ctx context.Context, client client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
	if c.listError != nil {
		return c.listError
	}
	c.listCount++
	return client.List(ctx, list, opts...)
}

func (c *callCounter) patchCall(ctx context.Context, client client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	if c.patchError != nil {
		return c.patchError
	}

	c.patchCount++
	if reflect.TypeOf(obj) == reflect.TypeFor[*corev1.PersistentVolumeClaim]() {
		pvc, ok := obj.(*corev1.PersistentVolumeClaim)
		if !ok {
			return assert.AnError
		}
		c.patchedPVCs = append(c.patchedPVCs, pvc.GetName())
		c.patchedPVCStorageQuantities = append(c.patchedPVCStorageQuantities, pvc.Spec.Resources.Requests.Storage())
	}

	if reflect.TypeOf(obj) == reflect.TypeFor[*monitoringv1.Prometheus]() {
		prom, ok := obj.(*monitoringv1.Prometheus)
		if !ok {
			return assert.AnError
		}

		if prom.Spec.Paused {
			c.patchedPromPaused = Paused
		} else {
			c.patchedPromPaused = NotPaused
		}

		c.patchedPromStorageQuantity = prom.Spec.Storage.VolumeClaimTemplate.Spec.Resources.Requests.Storage()
	}

	return client.Patch(ctx, obj, patch, opts...)
}

func (c *callCounter) deleteAllOfCall(ctx context.Context, k8sClient client.WithWatch, obj client.Object, opts ...client.DeleteAllOfOption) error {
	if c.deleteAllOfError != nil {
		return c.deleteAllOfError
	}

	if reflect.TypeOf(obj) == reflect.TypeFor[*appsv1.StatefulSet]() {
		c.deleteAllOfType = "StatefulSet"
	}

	deleteAllOfOpts := client.DeleteAllOfOptions{}
	deleteAllOfOpts.ApplyOptions(opts)
	c.deleteAllOfPropagationPolicy = *deleteAllOfOpts.AsDeleteOptions().PropagationPolicy
	c.deleteAllOfLabelSelector = deleteAllOfOpts.AsListOptions().LabelSelector
	c.deleteAllOfTypenNamespace = deleteAllOfOpts.ListOptions.Namespace

	return k8sClient.DeleteAllOf(ctx, obj, opts...)
}

func newFakeClientBuilder(t *testing.T) *fake.ClientBuilder {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, appsv1.AddToScheme(scheme))
	require.NoError(t, monitoringv1.AddToScheme(scheme))
	return fake.NewClientBuilder().WithScheme(scheme)
}

func newFakeClientBuilderWithCounter(t *testing.T, callCounter *callCounter) *fake.ClientBuilder {
	return newFakeClientBuilder(t).
		WithInterceptorFuncs(interceptor.Funcs{
			Get:         callCounter.getCall,
			List:        callCounter.listCall,
			Patch:       callCounter.patchCall,
			DeleteAllOf: callCounter.deleteAllOfCall,
		})
}

func newPVCListForTest(items ...corev1.PersistentVolumeClaim) *corev1.PersistentVolumeClaimList {
	return &corev1.PersistentVolumeClaimList{Items: items}
}

func newPVCForTest(t *testing.T, name string, storageSize string, labels map[string]string) corev1.PersistentVolumeClaim {
	storageQuantity, err := resource.ParseQuantity(storageSize)
	require.NoError(t, err)

	return corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsForTest,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.VolumeResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: storageQuantity,
				},
			},
		},
	}
}

func newPrometheusForTest(name string, storageSize string) *monitoringv1.Prometheus {
	return &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsForTest,
		},
		Spec: monitoringv1.PrometheusSpec{
			CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
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
		config: config{
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

func (f *fakeResizeSteps) patchPrometheus(ctx context.Context) error {
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
