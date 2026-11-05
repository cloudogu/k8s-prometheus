package storageresize

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestPatchPVCs(t *testing.T) {
	t.Run("should only patch PVCs that are part of prometheus", func(t *testing.T) {
		t.Skip("TODO")
		counter := &callCounter{}
		fakeClient := newFakeClientBuilderWithCounter(t, counter).Build()

		err := patchPVCs(fakeClient, "10G")

		assert.NoError(t, err)
	})
}

type callCounter struct {
	listCount int
}

func (c *callCounter) listCall(ctx context.Context, client client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
	c.listCount++
	return client.List(ctx, list, opts...)
}

func newFakeClientBuilder(t *testing.T) *fake.ClientBuilder {
	scheme := runtime.NewScheme()
	return fake.NewClientBuilder().WithScheme(scheme)
}

func newFakeClientBuilderWithCounter(t *testing.T, callCounter *callCounter) *fake.ClientBuilder {
	return newFakeClientBuilder(t).
		WithInterceptorFuncs(interceptor.Funcs{
			List: callCounter.listCall,
		})
}

func newPVCListForTest() *corev1.PersistentVolumeClaimList {
	return &corev1.PersistentVolumeClaimList{
		Items: []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns",
					Name:      "pvc1",
				},
				Spec: corev1.PersistentVolumeClaimSpec{},
			},
		},
	}
}
