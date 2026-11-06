package storageresize

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cloudogu/k8s-prometheus/auth/internal/logging"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ResizeStorage() error {
	cfg, err := readConfig()
	if err != nil {
		return fmt.Errorf("read config %w", err)
	}

	logging.ConfigureLogger(cfg.logLevel)

	k8sClientConfig := ctrl.GetConfigOrDie()
	k8sClient, err := client.New(k8sClientConfig, client.Options{})
	if err != nil {
		return fmt.Errorf("create k8s client: %w", err)
	}
	err = monitoringv1.AddToScheme(k8sClient.Scheme())
	if err != nil {
		return fmt.Errorf("register scheme for prometheus CR: %w", err)
	}

	ctx := context.Background()
	err = patchPVCs(ctx, k8sClient, cfg.namespace, cfg.promOperatorName, cfg.storageSize)
	if err != nil {
		return fmt.Errorf("patch PVCs: %w", err)
	}

	err = setPrometheusPaused(ctx, k8sClient, cfg.namespace, cfg.promOperatorName, true)
	if err != nil {
		return fmt.Errorf("pause prometheus reconciliation %w", err)
	}

	err = patchPrometheus(ctx, k8sClient, cfg.namespace, cfg.promOperatorName, cfg.storageSize)
	if err != nil {
		return fmt.Errorf("patch prometheus: %w", err)
	}

	err = deleteStatefulSet(ctx, k8sClient, cfg.namespace, cfg.promOperatorName)
	if err != nil {
		return fmt.Errorf("delete statefulset: %w", err)
	}

	err = setPrometheusPaused(ctx, k8sClient, cfg.namespace, cfg.promOperatorName, false)
	if err != nil {
		return fmt.Errorf("continue prometheus reconciliation %w", err)
	}

	return nil
}

func patchPVCs(ctx context.Context, k8sClient client.Client, namespace string, promOpName string, storageSize string) error {
	slog.Info(fmt.Sprintf("patch storage=%s of PVC's with label 'operator.prometheus.io/name=%s'", storageSize, promOpName))

	pvcList := &corev1.PersistentVolumeClaimList{}
	err := k8sClient.List(ctx, pvcList,
		client.InNamespace(namespace),
		client.MatchingLabels{"operator.prometheus.io/name": promOpName})
	if err != nil {
		return fmt.Errorf("get list of PVCs: %w", err)
	}

	storageQuantity, err := resource.ParseQuantity(storageSize)
	if err != nil {
		return fmt.Errorf("parse storage quantity '%s': %w", storageSize, err)
	}

	for _, pvc := range pvcList.Items {
		pvcBeforePatch := pvc.DeepCopy()
		pvc.Spec.Resources.Requests[corev1.ResourceStorage] = storageQuantity
		patchErr := k8sClient.Patch(ctx, &pvc, client.MergeFrom(pvcBeforePatch))
		if patchErr != nil {
			return fmt.Errorf("patch PVC '%s' storage to %s: %w", pvc.GetName(), storageSize, patchErr)
		}
	}

	return nil
}

func setPrometheusPaused(ctx context.Context, k8sClient client.Client, namespace string, promOpName string, pause bool) error {
	slog.Info(fmt.Sprintf("set prometheus resource spec.paused=%v", pause))

	prom := &monitoringv1.Prometheus{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: promOpName, Namespace: namespace}, prom)
	if err != nil {
		return fmt.Errorf("get prometheus resource: %w", err)
	}
	promBeforePatch := prom.DeepCopy()
	prom.Spec.Paused = pause
	patchErr := k8sClient.Patch(ctx, prom, client.MergeFrom(promBeforePatch))
	if patchErr != nil {
		return fmt.Errorf("patch prometheus resource spec.paused=%v,  %w", pause, patchErr)
	}

	return nil
}

func patchPrometheus(ctx context.Context, k8sClient client.Client, namespace string, promOpName string, storageSize string) error {
	slog.Info(fmt.Sprintf("patch prometheus resource spec.storage.volumeClaimTemplate.spec.resources.requests.storage=%s", storageSize))

	prom := &monitoringv1.Prometheus{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: promOpName, Namespace: namespace}, prom)
	if err != nil {
		return fmt.Errorf("get prometheus resource: %w", err)
	}

	storageQuantity, err := resource.ParseQuantity(storageSize)
	if err != nil {
		return fmt.Errorf("parse storage quantity '%s': %w", storageSize, err)
	}

	promBeforePatch := prom.DeepCopy()
	prom.Spec.Storage.VolumeClaimTemplate.Spec.Resources.Requests[corev1.ResourceStorage] = storageQuantity
	patchErr := k8sClient.Patch(ctx, prom, client.MergeFrom(promBeforePatch))
	if patchErr != nil {
		return fmt.Errorf(
			"patch prometheus resource spec.storage.volumeClaimTemplate.spec.resources.requests.storage=%s: %w",
			storageSize, patchErr,
		)
	}

	return nil
}

func deleteStatefulSet(ctx context.Context, k8sClient client.Client, namespace string, promOpName string) error {
	slog.Info(fmt.Sprintf("delete statefulset with label 'operator.prometheus.io/name=%s'", promOpName))

	statefulSet := &appsv1.StatefulSet{}
	err := k8sClient.DeleteAllOf(ctx, statefulSet,
		client.InNamespace(namespace),
		client.MatchingLabels{"operator.prometheus.io/name": promOpName},
		client.PropagationPolicy(metav1.DeletePropagationOrphan))
	if err != nil {
		return fmt.Errorf("delete statefulset: %w", err)
	}
	return nil
}
