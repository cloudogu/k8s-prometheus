package storageresize

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPatchPVCs(t *testing.T) {
	t.Run("should only patch PVCs that are part of prometheus", func(t *testing.T) {
		pvcList := newPVCListForTest(
			newPVCForTest(t, "pvc1", "10Gi", map[string]string{
				"operator.prometheus.io/name": promOpNameForTest,
			}),
			newPVCForTest(t, "pvc2", "10Gi", map[string]string{
				"operator.prometheus.io/name": promOpNameForTest,
			}),
			newPVCForTest(t, "pvc3", "5Gi", map[string]string{
				"operator.prometheus.io/name": "not_the_right_prom_op_name",
			}),
			newPVCForTest(t, "pvc4", "5Gi", map[string]string{
				"example.com/name": "aValue",
			}),
		)
		counter := &callCounter{}
		fakeClient := newFakeClientBuilderWithCounter(t, counter).
			WithLists(pvcList).
			Build()

		err := patchPVCs(context.Background(), fakeClient, nsForTest, promOpNameForTest, "20Gi")

		assert.NoError(t, err)
		assert.Equal(t, 1, counter.listCount)
		assert.Equal(t, 2, counter.patchCount)
		assert.Equal(t, []string{"pvc1", "pvc2"}, counter.patchedPVCs)

		quantitiesAreEqual := slices.EqualFunc(
			counter.patchedPVCStorageQuantities,
			[]resource.Quantity{
				resource.MustParse("20Gi"),
				resource.MustParse("20Gi"),
			},
			func(a *resource.Quantity, b resource.Quantity) bool { return a.Cmp(b) == 0 },
		)
		assert.True(t, quantitiesAreEqual)
	})

	t.Run("should propagate a client list error", func(t *testing.T) {
		pvcList := newPVCListForTest()
		counter := &callCounter{listError: errors.New("listError")}
		fakeClient := newFakeClientBuilderWithCounter(t, counter).
			WithLists(pvcList).
			Build()

		err := patchPVCs(context.Background(), fakeClient, nsForTest, promOpNameForTest, "20Gi")

		assert.Error(t, err)
		assert.ErrorContains(t, err, "listError")
	})

	t.Run("should propagate a client patch error", func(t *testing.T) {
		pvcList := newPVCListForTest(
			newPVCForTest(t, "pvc1", "10Gi", map[string]string{
				"operator.prometheus.io/name": promOpNameForTest,
			}),
		)
		counter := &callCounter{patchError: errors.New("patchError")}
		fakeClient := newFakeClientBuilderWithCounter(t, counter).
			WithLists(pvcList).
			Build()

		err := patchPVCs(context.Background(), fakeClient, nsForTest, promOpNameForTest, "20Gi")

		assert.Error(t, err)
		assert.ErrorContains(t, err, "patchError")
	})
}

func TestSetPrometheusPaused(t *testing.T) {
	t.Run("should pause reconciliation of the prometheus resource", func(t *testing.T) {
		prom := newPrometheusForTest(promOpNameForTest, "10Gi")
		counter := &callCounter{patchedPromPaused: Unknown}
		fakeClient := newFakeClientBuilderWithCounter(t, counter).
			WithObjects(prom).
			Build()

		err := setPrometheusPaused(context.Background(), fakeClient, nsForTest, promOpNameForTest, true)

		assert.NoError(t, err)
		assert.Equal(t, 1, counter.getCount)
		assert.Equal(t, 1, counter.patchCount)
		assert.Equal(t, Paused, counter.patchedPromPaused)
	})

	t.Run("should continue reconciliation of the prometheus resource", func(t *testing.T) {
		prom := newPrometheusForTest(promOpNameForTest, "10Gi")
		counter := &callCounter{patchedPromPaused: Unknown}
		fakeClient := newFakeClientBuilderWithCounter(t, counter).
			WithObjects(prom).
			Build()

		err := setPrometheusPaused(context.Background(), fakeClient, nsForTest, promOpNameForTest, false)

		assert.NoError(t, err)
		assert.Equal(t, 1, counter.getCount)
		assert.Equal(t, 1, counter.patchCount)
		assert.Equal(t, NotPaused, counter.patchedPromPaused)
	})

	t.Run("should propagate a client get error", func(t *testing.T) {
		prom := newPrometheusForTest(promOpNameForTest, "10Gi")
		counter := &callCounter{getError: errors.New("getError")}
		fakeClient := newFakeClientBuilderWithCounter(t, counter).
			WithObjects(prom).
			Build()

		err := setPrometheusPaused(context.Background(), fakeClient, nsForTest, promOpNameForTest, false)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "getError")
	})

	t.Run("should propagate a client patch error", func(t *testing.T) {
		prom := newPrometheusForTest(promOpNameForTest, "10Gi")
		counter := &callCounter{patchError: errors.New("patchError")}
		fakeClient := newFakeClientBuilderWithCounter(t, counter).
			WithObjects(prom).
			Build()

		err := setPrometheusPaused(context.Background(), fakeClient, nsForTest, promOpNameForTest, false)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "patchError")
	})
}

func TestPatchPrometheus(t *testing.T) {
	t.Run("should patch prometheus resources", func(t *testing.T) {
		prom := newPrometheusForTest(promOpNameForTest, "10Gi")
		counter := &callCounter{patchedPromPaused: Unknown}
		fakeClient := newFakeClientBuilderWithCounter(t, counter).
			WithObjects(prom).
			Build()

		err := patchPrometheus(context.Background(), fakeClient, nsForTest, promOpNameForTest, "12Gi")

		assert.NoError(t, err)
		assert.Equal(t, 1, counter.getCount)
		assert.Equal(t, 1, counter.patchCount)
		assert.True(t, counter.patchedPromStorageQuantity.Cmp(resource.MustParse("12Gi")) == 0)
	})
}

func TestDeleteStatefulSet(t *testing.T) {
	t.Run("should delete statefulset that is part of prometheus with propagation policy orphan", func(t *testing.T) {
		counter := &callCounter{}
		fakeClient := newFakeClientBuilderWithCounter(t, counter).Build()

		err := deleteStatefulSet(context.Background(), fakeClient, nsForTest, promOpNameForTest)

		assert.NoError(t, err)
		assert.Equal(t, "StatefulSet", counter.deleteAllOfType)
		assert.Equal(t, nsForTest, counter.deleteAllOfTypenNamespace)
		assert.Equal(t, metav1.DeletePropagationOrphan, counter.deleteAllOfPropagationPolicy)
		assert.Equal(t, fmt.Sprintf("operator.prometheus.io/name=%s", promOpNameForTest), counter.deleteAllOfLabelSelector)
	})

	t.Run("should propagate a client deleteAllOf error", func(t *testing.T) {
		counter := &callCounter{deleteAllOfError: errors.New("deleteAllOfError")}
		fakeClient := newFakeClientBuilderWithCounter(t, counter).Build()

		err := deleteStatefulSet(context.Background(), fakeClient, nsForTest, promOpNameForTest)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "deleteAllOfError")
	})
}
