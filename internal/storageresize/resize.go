package storageresize

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const prometheusOperatorNameLabel = "operator.prometheus.io/name"

type resizeSteps interface {
	patchPVCs(ctx context.Context) error
	setPrometheusPaused(ctx context.Context, pause bool) error
	patchPrometheusVolumeClaimTemplate(ctx context.Context) error
	deleteStatefulSet(ctx context.Context) error
}

type defaultResizeSteps struct {
	config    *Config
	k8sClient client.Client
}

func NewResizeSteps(config *Config, k8sClient client.Client) *defaultResizeSteps {
	return &defaultResizeSteps{
		config:    config,
		k8sClient: k8sClient,
	}
}

// CreateK8sClient reads the Kubernetes configuration and creates the k8s client. It also adds the schema
// used by Prometheus to the k8s client.
func CreateK8sClient() (client.Client, error) {
	k8sClientConfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("get config for the kubernetes api server: %w", err)
	}
	k8sClient, err := client.New(k8sClientConfig, client.Options{})
	if err != nil {
		return nil, fmt.Errorf("create k8s client: %w", err)
	}
	err = monitoringv1.AddToScheme(k8sClient.Scheme())
	if err != nil {
		return nil, fmt.Errorf("register scheme for prometheus CR: %w", err)
	}
	return k8sClient, nil
}

func (d *defaultResizeSteps) patchPVCs(ctx context.Context) error {
	slog.Info("patch storage size of PVC's", "size", d.config.storageSize, prometheusOperatorNameLabel, d.config.promOperatorName)

	pvcList := &corev1.PersistentVolumeClaimList{}
	err := d.k8sClient.List(ctx, pvcList,
		client.InNamespace(d.config.namespace),
		client.MatchingLabels{prometheusOperatorNameLabel: d.config.promOperatorName})
	if err != nil {
		return fmt.Errorf("list PVCs: %w", err)
	}

	storageQuantity, err := resource.ParseQuantity(d.config.storageSize)
	if err != nil {
		return fmt.Errorf("parse storage quantity %q: %w", d.config.storageSize, err)
	}

	for _, pvc := range pvcList.Items {
		pvcBeforePatch := pvc.DeepCopy()
		pvc.Spec.Resources.Requests[corev1.ResourceStorage] = storageQuantity
		patchErr := d.k8sClient.Patch(ctx, &pvc, client.MergeFrom(pvcBeforePatch))
		if patchErr != nil {
			return fmt.Errorf("patch PVC %q storage to %q: %w", pvc.GetName(), d.config.storageSize, patchErr)
		}
	}

	return nil
}

func (d *defaultResizeSteps) setPrometheusPaused(ctx context.Context, pause bool) error {
	slog.Info("set prometheus resource paused", "spec.paused", pause)

	prom := &monitoringv1.Prometheus{}
	err := d.k8sClient.Get(ctx, types.NamespacedName{Name: d.config.promOperatorName, Namespace: d.config.namespace}, prom)
	if err != nil {
		return fmt.Errorf("get prometheus resource: %w", err)
	}
	promBeforePatch := prom.DeepCopy()
	prom.Spec.Paused = pause
	err = d.k8sClient.Patch(ctx, prom, client.MergeFrom(promBeforePatch))
	if err != nil {
		return fmt.Errorf("patch prometheus resource spec.paused=%t:  %w", pause, err)
	}

	return nil
}

func (d *defaultResizeSteps) patchPrometheusVolumeClaimTemplate(ctx context.Context) error {
	slog.Info("patch prometheus VolumeClaimTemplate", "size", d.config.storageSize)

	prom := &monitoringv1.Prometheus{}
	err := d.k8sClient.Get(ctx, types.NamespacedName{Name: d.config.promOperatorName, Namespace: d.config.namespace}, prom)
	if err != nil {
		return fmt.Errorf("get prometheus resource: %w", err)
	}

	storageQuantity, err := resource.ParseQuantity(d.config.storageSize)
	if err != nil {
		return fmt.Errorf("parse storage quantity %q: %w", d.config.storageSize, err)
	}

	promBeforePatch := prom.DeepCopy()
	prom.Spec.Storage.VolumeClaimTemplate.Spec.Resources.Requests[corev1.ResourceStorage] = storageQuantity
	err = d.k8sClient.Patch(ctx, prom, client.MergeFrom(promBeforePatch))
	if err != nil {
		return fmt.Errorf("patch prometheus VolumeClaimTemplate size %q: %w", d.config.storageSize, err)
	}

	return nil
}

func (d *defaultResizeSteps) deleteStatefulSet(ctx context.Context) error {
	slog.Info("delete statefulset", prometheusOperatorNameLabel, d.config.promOperatorName)

	statefulSet := &appsv1.StatefulSet{}
	err := d.k8sClient.DeleteAllOf(ctx, statefulSet,
		client.InNamespace(d.config.namespace),
		client.MatchingLabels{prometheusOperatorNameLabel: d.config.promOperatorName},
		client.PropagationPolicy(metav1.DeletePropagationOrphan))
	if err != nil {
		return fmt.Errorf("delete all statefulsets that belong to prometheus: %w", err)
	}
	return nil
}

func ResizeStorage(resizeSteps resizeSteps) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := resizeSteps.patchPVCs(ctx)
	if err != nil {
		return fmt.Errorf("patch PVCs: %w", err)
	}

	err = resizeSteps.setPrometheusPaused(ctx, true)
	if err != nil {
		return fmt.Errorf("pause prometheus reconciliation: %w", err)
	}

	err = resizeSteps.patchPrometheusVolumeClaimTemplate(ctx)
	if err != nil {
		return fmt.Errorf("patch prometheus VolumeClaimTemplate: %w", err)
	}

	err = resizeSteps.deleteStatefulSet(ctx)
	if err != nil {
		return fmt.Errorf("delete statefulset: %w", err)
	}

	err = resizeSteps.setPrometheusPaused(ctx, false)
	if err != nil {
		return fmt.Errorf("continue prometheus reconciliation: %w", err)
	}

	return nil
}
