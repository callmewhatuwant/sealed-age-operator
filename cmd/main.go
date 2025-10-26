// cmd/main.go
/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
*/
package main

import (
	"flag"
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	securityv1alpha1 "github.com/callmewhatuwant/sealed-age-operator/api/v1alpha1"
	"github.com/callmewhatuwant/sealed-age-operator/internal/controller"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(securityv1alpha1.AddToScheme(scheme))
}

func main() {
	var (
		metricsAddr          string
		probeAddr            string
		enableLeaderElection bool
		leaderNS             string

		// key Secret Discovery
		keyNS, keyLabelKey, keyLabelVal string
	)

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metrics endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager.")
	flag.StringVar(&leaderNS, "leader-election-namespace", "", "Namespace for the leader election Lease (defaults to POD_NAMESPACE or sealed-age-system).")

	flag.StringVar(&keyNS, "key-namespace", "sealed-age-system", "Namespace containing AGE key Secrets.")
	flag.StringVar(&keyLabelKey, "key-label-key", "app", "Label key for AGE key Secrets.")
	flag.StringVar(&keyLabelVal, "key-label-val", "age-key", "Label value for AGE key Secrets.")

	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	// Resolve leader election namespace:
	// 1) explicit flag > 2) POD_NAMESPACE > 3) sealed-age-system
	if leaderNS == "" {
		if podNS := os.Getenv("POD_NAMESPACE"); podNS != "" {
			leaderNS = podNS
		} else {
			leaderNS = "sealed-age-system"
		}
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,

		// 🔒 Leader Election (Lease in coordination.k8s.io)
		LeaderElection:          enableLeaderElection,
		LeaderElectionID:        "sealed-age-operator.age.io",
		LeaderElectionNamespace: leaderNS,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := (&controller.SealedAgeReconciler{
		Client:       mgr.GetClient(),
		Scheme:       mgr.GetScheme(),
		KeyNamespace: keyNS,
		KeyLabelKey:  keyLabelKey,
		KeyLabelVal:  keyLabelVal,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SealedAge")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager", "leaderElection", enableLeaderElection, "leaderNamespace", leaderNS)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
