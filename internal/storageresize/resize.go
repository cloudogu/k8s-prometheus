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

type resizeSteps interface {
	configure() error
	createK8sClient() error
	patchPVCs(ctx context.Context) error
	setPrometheusPaused(ctx context.Context, pause bool) error
	patchPrometheus(ctx context.Context) error
	deleteStatefulSet(ctx context.Context) error
}

type defaultResizeSteps struct {
	config    config
	k8sClient client.Client
}

func NewResizeSteps() *defaultResizeSteps {
	return &defaultResizeSteps{}
}

func (d *defaultResizeSteps) configure() error {
	cfg, err := readConfig()
	if err != nil {
		return fmt.Errorf("read config %w", err)
	}
	logging.ConfigureLogger(cfg.logLevel)
	d.config = *cfg
	return nil
}

func (d *defaultResizeSteps) createK8sClient() error {
	k8sClientConfig := ctrl.GetConfigOrDie()
	k8sClient, err := client.New(k8sClientConfig, client.Options{})
	if err != nil {
		return fmt.Errorf("create k8s client: %w", err)
	}
	err = monitoringv1.AddToScheme(k8sClient.Scheme())
	if err != nil {
		return fmt.Errorf("register scheme for prometheus CR: %w", err)
	}
	d.k8sClient = k8sClient
	return nil
}

func (d *defaultResizeSteps) patchPVCs(ctx context.Context) error {
	slog.Info(fmt.Sprintf("patch storage=%s of PVC's with label 'operator.prometheus.io/name=%s'", d.config.storageSize, d.config.promOperatorName))

	pvcList := &corev1.PersistentVolumeClaimList{}
	err := d.k8sClient.List(ctx, pvcList,
		client.InNamespace(d.config.namespace),
		client.MatchingLabels{"operator.prometheus.io/name": d.config.promOperatorName})
	if err != nil {
		return fmt.Errorf("get list of PVCs: %w", err)
	}

	storageQuantity, err := resource.ParseQuantity(d.config.storageSize)
	if err != nil {
		return fmt.Errorf("parse storage quantity '%s': %w", d.config.storageSize, err)
	}

	for _, pvc := range pvcList.Items {
		pvcBeforePatch := pvc.DeepCopy()
		pvc.Spec.Resources.Requests[corev1.ResourceStorage] = storageQuantity
		patchErr := d.k8sClient.Patch(ctx, &pvc, client.MergeFrom(pvcBeforePatch))
		if patchErr != nil {
			return fmt.Errorf("patch PVC '%s' storage to %s: %w", pvc.GetName(), d.config.storageSize, patchErr)
		}
	}

	return nil
}

func (d *defaultResizeSteps) setPrometheusPaused(ctx context.Context, pause bool) error {
	slog.Info(fmt.Sprintf("set prometheus resource spec.paused=%v", pause))

	prom := &monitoringv1.Prometheus{}
	err := d.k8sClient.Get(ctx, types.NamespacedName{Name: d.config.promOperatorName, Namespace: d.config.namespace}, prom)
	if err != nil {
		return fmt.Errorf("get prometheus resource: %w", err)
	}
	promBeforePatch := prom.DeepCopy()
	prom.Spec.Paused = pause
	err = d.k8sClient.Patch(ctx, prom, client.MergeFrom(promBeforePatch))
	if err != nil {
		return fmt.Errorf("patch prometheus resource spec.paused=%v,  %w", pause, err)
	}

	return nil
}

func (d *defaultResizeSteps) patchPrometheus(ctx context.Context) error {
	slog.Info(fmt.Sprintf("patch prometheus resource spec.storage.volumeClaimTemplate.spec.resources.requests.storage=%s", d.config.storageSize))

	prom := &monitoringv1.Prometheus{}
	err := d.k8sClient.Get(ctx, types.NamespacedName{Name: d.config.promOperatorName, Namespace: d.config.namespace}, prom)
	if err != nil {
		return fmt.Errorf("get prometheus resource: %w", err)
	}

	storageQuantity, err := resource.ParseQuantity(d.config.storageSize)
	if err != nil {
		return fmt.Errorf("parse storage quantity '%s': %w", d.config.storageSize, err)
	}

	promBeforePatch := prom.DeepCopy()
	prom.Spec.Storage.VolumeClaimTemplate.Spec.Resources.Requests[corev1.ResourceStorage] = storageQuantity
	err = d.k8sClient.Patch(ctx, prom, client.MergeFrom(promBeforePatch))
	if err != nil {
		return fmt.Errorf(
			"patch prometheus resource spec.storage.volumeClaimTemplate.spec.resources.requests.storage=%s: %w",
			d.config.storageSize, err,
		)
	}

	return nil
}

func (d *defaultResizeSteps) deleteStatefulSet(ctx context.Context) error {
	slog.Info(fmt.Sprintf("delete statefulset with label 'operator.prometheus.io/name=%s'", d.config.promOperatorName))

	statefulSet := &appsv1.StatefulSet{}
	err := d.k8sClient.DeleteAllOf(ctx, statefulSet,
		client.InNamespace(d.config.namespace),
		client.MatchingLabels{"operator.prometheus.io/name": d.config.promOperatorName},
		client.PropagationPolicy(metav1.DeletePropagationOrphan))
	if err != nil {
		return fmt.Errorf("delete statefulset: %w", err)
	}
	return nil
}

func ResizeStorage(resizeSteps resizeSteps) error {
	err := resizeSteps.configure()
	if err != nil {
		return fmt.Errorf("configure %w", err)
	}

	err = resizeSteps.createK8sClient()
	if err != nil {
		return fmt.Errorf("create k8s client: %w", err)
	}

	ctx := context.Background()
	err = resizeSteps.patchPVCs(ctx)
	if err != nil {
		return fmt.Errorf("patch PVCs: %w", err)
	}

	err = resizeSteps.setPrometheusPaused(ctx, true)
	if err != nil {
		return fmt.Errorf("pause prometheus reconciliation %w", err)
	}

	err = resizeSteps.patchPrometheus(ctx)
	if err != nil {
		return fmt.Errorf("patch prometheus: %w", err)
	}

	err = resizeSteps.deleteStatefulSet(ctx)
	if err != nil {
		return fmt.Errorf("delete statefulset: %w", err)
	}

	err = resizeSteps.setPrometheusPaused(ctx, false)
	if err != nil {
		return fmt.Errorf("continue prometheus reconciliation %w", err)
	}

	return nil
}
