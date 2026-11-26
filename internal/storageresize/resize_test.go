package storageresize

import (
	"context"
	"errors"
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

var testCtx = context.Background()

func TestNewResizeSteps(t *testing.T) {
	t.Run("should set the fields correctly", func(t *testing.T) {
		cfg := &Config{}
		k8sClient := fake.NewFakeClient()

		sut := NewResizeSteps(cfg, k8sClient)

		assert.Same(t, cfg, sut.config)
		assert.Same(t, k8sClient, sut.k8sClient)
	})
}

func TestPatchPVCs(t *testing.T) {
	t.Run("should only patch those two PVCs that are part of prometheus", func(t *testing.T) {
		pvcList := &corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{
			newPVCForTest(t, "pvc1", "10Gi", map[string]string{
				prometheusOperatorNameLabel: promOpNameForTest,
			}),
			newPVCForTest(t, "pvc2", "10Gi", map[string]string{
				prometheusOperatorNameLabel: promOpNameForTest,
			}),
			newPVCForTest(t, "pvc3", "5Gi", map[string]string{
				prometheusOperatorNameLabel: "not_the_right_prom_op_name",
			}),
			newPVCForTest(t, "pvc4", "5Gi", map[string]string{
				"example.com/name": "aValue",
			}),
		}}
		var listWasCalled bool
		var actualPatchedPvcs []string
		var actualPvcQuantities []resource.Quantity
		fakeClient := newFakeClientBuilder(t).
			WithLists(pvcList).
			WithInterceptorFuncs(interceptor.Funcs{
				List: func(ctx context.Context, client client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
					listWasCalled = true
					return client.List(ctx, list, opts...)
				},
				Patch: func(ctx context.Context, client client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
					pvc, ok := obj.(*corev1.PersistentVolumeClaim)
					require.True(t, ok)
					actualPatchedPvcs = append(actualPatchedPvcs, pvc.GetName())
					actualPvcQuantities = append(actualPvcQuantities, *pvc.Spec.Resources.Requests.Storage())
					return client.Patch(ctx, obj, patch, opts...)
				},
			}).
			Build()
		sut := newResizeStepsForTest(fakeClient, "20Gi")

		err := sut.patchPVCs(testCtx)

		require.NoError(t, err)
		assert.True(t, listWasCalled)
		assert.Equal(t, []string{"pvc1", "pvc2"}, actualPatchedPvcs)

		expectedPvcQuantities := []resource.Quantity{
			resource.MustParse("20Gi"),
			resource.MustParse("20Gi"),
		}
		assert.ElementsMatch(t, expectedPvcQuantities, actualPvcQuantities)
	})

	t.Run("should propagate a client list error", func(t *testing.T) {
		pvcList := &corev1.PersistentVolumeClaimList{}
		fakeClient := newFakeClientBuilder(t).
			WithLists(pvcList).
			WithInterceptorFuncs(interceptor.Funcs{
				List: func(ctx context.Context, client client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
					return assert.AnError
				},
			}).
			Build()
		sut := newResizeStepsForTest(fakeClient, "20Gi")

		err := sut.patchPVCs(testCtx)

		assert.ErrorContains(t, err, "list PVCs")
	})

	t.Run("should propagate a quantity parse error", func(t *testing.T) {
		pvcList := &corev1.PersistentVolumeClaimList{}
		fakeClient := newFakeClientBuilder(t).
			WithLists(pvcList).
			Build()
		sut := newResizeStepsForTest(fakeClient, "XXGi")

		err := sut.patchPVCs(testCtx)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "storage")
	})

	t.Run("should propagate a client patch error", func(t *testing.T) {
		pvcList := &corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{
			newPVCForTest(t, "pvc1", "10Gi", map[string]string{
				prometheusOperatorNameLabel: promOpNameForTest,
			}),
		}}
		fakeClient := newFakeClientBuilder(t).
			WithLists(pvcList).
			WithInterceptorFuncs(interceptor.Funcs{
				Patch: func(ctx context.Context, client client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
					return assert.AnError
				},
			}).
			Build()
		sut := newResizeStepsForTest(fakeClient, "20Gi")

		err := sut.patchPVCs(testCtx)

		assert.ErrorContains(t, err, "patch PVC")
	})
}

func TestSetPrometheusPaused(t *testing.T) {
	t.Run("should pause reconciliation of the prometheus resource", func(t *testing.T) {
		prom := newPrometheusForTest(promOpNameForTest, "10Gi", false)
		var getCalled bool
		var patchCalled bool
		fakeClient := newFakeClientBuilder(t).
			WithObjects(prom).
			WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					getCalled = true
					return client.Get(ctx, key, obj, opts...)
				},
				Patch: func(ctx context.Context, client client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
					patchCalled = true
					patchedProm, ok := obj.(*monitoringv1.Prometheus)
					require.True(t, ok)
					assert.True(t, patchedProm.Spec.Paused)
					return client.Patch(ctx, obj, patch, opts...)
				},
			}).
			Build()
		sut := newResizeStepsForTest(fakeClient, "20Gi")

		err := sut.setPrometheusPaused(testCtx, true)

		require.NoError(t, err)
		assert.True(t, getCalled)
		assert.True(t, patchCalled)
	})

	t.Run("should continue reconciliation of the prometheus resource", func(t *testing.T) {
		prom := newPrometheusForTest(promOpNameForTest, "10Gi", true)
		var getCalled bool
		var patchCalled bool
		fakeClient := newFakeClientBuilder(t).
			WithObjects(prom).
			WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					getCalled = true
					return client.Get(ctx, key, obj, opts...)
				},
				Patch: func(ctx context.Context, client client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
					patchCalled = true
					patchedProm, ok := obj.(*monitoringv1.Prometheus)
					require.True(t, ok)
					assert.False(t, patchedProm.Spec.Paused)
					return client.Patch(ctx, obj, patch, opts...)
				},
			}).
			Build()
		sut := newResizeStepsForTest(fakeClient, "20Gi")

		err := sut.setPrometheusPaused(testCtx, false)

		require.NoError(t, err)
		assert.True(t, getCalled)
		assert.True(t, patchCalled)
	})

	t.Run("should propagate a client get error", func(t *testing.T) {
		prom := newPrometheusForTest(promOpNameForTest, "10Gi", false)
		fakeClient := newFakeClientBuilder(t).
			WithObjects(prom).
			WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					return assert.AnError
				},
			}).
			Build()
		sut := newResizeStepsForTest(fakeClient, "20Gi")

		err := sut.setPrometheusPaused(testCtx, false)

		assert.ErrorContains(t, err, "get prometheus")
	})

	t.Run("should propagate a client patch error", func(t *testing.T) {
		prom := newPrometheusForTest(promOpNameForTest, "10Gi", false)
		fakeClient := newFakeClientBuilder(t).
			WithObjects(prom).
			WithInterceptorFuncs(interceptor.Funcs{
				Patch: func(ctx context.Context, client client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
					return assert.AnError
				},
			}).
			Build()
		sut := newResizeStepsForTest(fakeClient, "20Gi")

		err := sut.setPrometheusPaused(testCtx, false)

		assert.ErrorContains(t, err, "patch prometheus")
	})
}

func TestPatchPrometheus(t *testing.T) {
	t.Run("should patch prometheus resources", func(t *testing.T) {
		prom := newPrometheusForTest(promOpNameForTest, "10Gi", false)
		var patchCalled bool
		fakeClient := newFakeClientBuilder(t).
			WithObjects(prom).
			WithInterceptorFuncs(interceptor.Funcs{
				Patch: func(ctx context.Context, client client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
					patchCalled = true
					patchedProm, ok := obj.(*monitoringv1.Prometheus)
					require.True(t, ok)
					assert.Equal(t, resource.MustParse("12Gi"), *patchedProm.Spec.Storage.VolumeClaimTemplate.Spec.Resources.Requests.Storage())
					return client.Patch(ctx, obj, patch, opts...)
				},
			}).
			Build()
		sut := newResizeStepsForTest(fakeClient, "12Gi")

		err := sut.patchPrometheusVolumeClaimTemplate(testCtx)

		require.NoError(t, err)
		assert.True(t, patchCalled)
	})

	t.Run("should propagate a client get error", func(t *testing.T) {
		prom := newPrometheusForTest(promOpNameForTest, "10Gi", false)
		var getCalled bool
		fakeClient := newFakeClientBuilder(t).
			WithObjects(prom).
			WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					getCalled = true
					return assert.AnError
				},
			}).
			Build()
		sut := newResizeStepsForTest(fakeClient, "12Gi")

		err := sut.patchPrometheusVolumeClaimTemplate(testCtx)

		assert.True(t, getCalled)
		assert.ErrorContains(t, err, "get prometheus")
	})

	t.Run("should propagate a quantity parse error", func(t *testing.T) {
		prom := newPrometheusForTest(promOpNameForTest, "10Gi", false)
		fakeClient := newFakeClientBuilder(t).
			WithObjects(prom).
			Build()
		sut := newResizeStepsForTest(fakeClient, "XXGi")

		err := sut.patchPrometheusVolumeClaimTemplate(testCtx)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "parse storage quantity")
	})

	t.Run("should propagate a client patch error", func(t *testing.T) {
		prom := newPrometheusForTest(promOpNameForTest, "10Gi", false)
		var patchCalled bool
		fakeClient := newFakeClientBuilder(t).
			WithObjects(prom).
			WithInterceptorFuncs(interceptor.Funcs{
				Patch: func(ctx context.Context, client client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
					patchCalled = true
					return assert.AnError
				},
			}).
			Build()
		sut := newResizeStepsForTest(fakeClient, "12Gi")

		err := sut.patchPrometheusVolumeClaimTemplate(testCtx)

		assert.True(t, patchCalled)
		assert.ErrorContains(t, err, "patch prometheus")
	})
}

func TestDeleteStatefulSet(t *testing.T) {
	t.Run("should delete statefulset that is part of prometheus with propagation policy orphan", func(t *testing.T) {
		var deleteAllOfCalled bool
		fakeClient := newFakeClientBuilder(t).
			WithInterceptorFuncs(interceptor.Funcs{
				DeleteAllOf: func(ctx context.Context, theClient client.WithWatch, obj client.Object, opts ...client.DeleteAllOfOption) error {
					deleteAllOfOpts := client.DeleteAllOfOptions{}
					deleteAllOfOpts.ApplyOptions(opts)

					deleteAllOfCalled = true
					assert.IsType(t, obj, &appsv1.StatefulSet{})
					assert.Equal(t, nsForTest, deleteAllOfOpts.ListOptions.Namespace)
					assert.Equal(t, prometheusOperatorNameLabel+"="+promOpNameForTest, deleteAllOfOpts.AsListOptions().LabelSelector)
					assert.Equal(t, metav1.DeletePropagationOrphan, *deleteAllOfOpts.AsDeleteOptions().PropagationPolicy)

					return theClient.DeleteAllOf(ctx, obj, opts...)
				},
			}).
			Build()
		sut := newResizeStepsForTest(fakeClient, "12Gi")

		err := sut.deleteStatefulSet(testCtx)

		require.NoError(t, err)
		assert.True(t, deleteAllOfCalled)
	})

	t.Run("should propagate a client deleteAllOf error", func(t *testing.T) {
		fakeClient := newFakeClientBuilder(t).
			WithInterceptorFuncs(interceptor.Funcs{
				DeleteAllOf: func(ctx context.Context, theClient client.WithWatch, obj client.Object, opts ...client.DeleteAllOfOption) error {
					return assert.AnError
				},
			}).
			Build()
		sut := newResizeStepsForTest(fakeClient, "12Gi")

		err := sut.deleteStatefulSet(testCtx)

		assert.ErrorContains(t, err, "delete all statefulsets")
	})
}

func TestResizeStorage(t *testing.T) {
	t.Run("should propagate patch PVC error", func(t *testing.T) {
		fakeResizeStps := &fakeResizeSteps{patchPVCError: errors.New("patchPVCErr")}

		err := ResizeStorage(fakeResizeStps)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "patchPVCErr")
	})

	t.Run("should propagate prometheus pause=true error", func(t *testing.T) {
		fakeResizeStps := &fakeResizeSteps{prometheusPauseErrors: []error{errors.New("firstPrometheusPauseErr")}}

		err := ResizeStorage(fakeResizeStps)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "firstPrometheusPauseErr")
	})

	t.Run("should propagate prometheus patch error", func(t *testing.T) {
		fakeResizeStps := &fakeResizeSteps{
			prometheusPauseErrors: []error{nil},
			patchPrometheusError:  errors.New("patchPrometheusError"),
		}

		err := ResizeStorage(fakeResizeStps)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "patchPrometheusError")
	})

	t.Run("should propagate delete statefulset error", func(t *testing.T) {
		fakeResizeStps := &fakeResizeSteps{
			prometheusPauseErrors:  []error{nil},
			deleteStatefulSetError: errors.New("deleteStatefulSetErr"),
		}

		err := ResizeStorage(fakeResizeStps)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "deleteStatefulSetErr")
	})

	t.Run("should propagate prometheus pause=false error", func(t *testing.T) {
		fakeResizeStps := &fakeResizeSteps{
			prometheusPauseErrors: []error{nil, errors.New("secondPrometheusPauseErr")},
		}

		err := ResizeStorage(fakeResizeStps)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "secondPrometheusPauseErr")
	})
}
