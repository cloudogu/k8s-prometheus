package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const errorFormat = "environment variable %s is not set"
const logLevelEnv = "LOG_LEVEL"
const promOpNameEnv = "PROM_OP_NAME"

func main() {
	logLevel := os.Getenv(logLevelEnv)
	if logLevel == "" {
		logLevel = "INFO"
	}

	promOpName := os.Getenv(promOpNameEnv)
	if promOpName == "" {
		panic(fmt.Sprintf(errorFormat, promOpNameEnv))
	}

	configureLogger(logLevel)

	slog.Info(fmt.Sprintf("l: %s, pn: %s", logLevel, promOpName))

	config := ctrl.GetConfigOrDie()
	k8sClient, err := client.New(config, client.Options{})
	if err != nil {
		panic(err.Error())
	}

	storageSize, err := resource.ParseQuantity("10Gi")
	if err != nil {
		panic(err.Error())
	}
	slog.Info("quantity: %v", storageSize)

	err = monitoringv1.AddToScheme(k8sClient.Scheme())
	if err != nil {
		panic(err.Error())
	}

	ctx := context.Background()
	prom := &monitoringv1.Prometheus{}
	err = k8sClient.Get(ctx, types.NamespacedName{Name: promOpName, Namespace: "ecosystem"}, prom)
	if err != nil {
		panic(err.Error())
	}
	slog.Info(fmt.Sprintf("prom: %s", prom.Name))

	pvcList := &corev1.PersistentVolumeClaimList{}
	err = k8sClient.List(ctx, pvcList,
		client.InNamespace("ecosystem"),
		client.MatchingLabels{"operator.prometheus.io/name": promOpName})
	if err != nil {
		panic(err.Error())
	}

	for _, pvc := range pvcList.Items {
		slog.Info(fmt.Sprintf("pvc: %s", pvc.Name))
	}

	/*
		statefulSet := &appsv1.StatefulSet{}
		err = k8sClient.DeleteAllOf(ctx, statefulSet,
			client.InNamespace("ecosystem"),
			client.MatchingLabels{"operator.prometheus.io/name": promOpName},
			client.PropagationPolicy(metav1.DeletePropagationOrphan))
		if err != nil {
			panic(err.Error())
		}

	*/
}

func configureLogger(logLevel string) {
	var level slog.Level
	var err = level.UnmarshalText([]byte(logLevel))
	if err != nil {
		slog.Error("error parsing log level. Setting log level to INFO.", "err", err)
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: false,
		Level:     level,
	}))
	slog.SetDefault(logger)

	slog.Info("configured logger", "level", level.String())
}
