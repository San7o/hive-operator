/*
                    GNU GENERAL PUBLIC LICENSE
                       Version 2, June 1991

 Copyright (C) 1989, 1991 Free Software Foundation, Inc.,
 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.
*/

// SPDX-License-Identifier: GPL-2.0-only

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"os"
	"strings"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
	"github.com/San7o/hive-operator/internal/controller"
	hiveDiscover "github.com/San7o/hive-operator/internal/controller"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(hivev1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var hiveProbeAddr string
	var hiveDataProbeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	var tlsOpts []func(*tls.Config)
	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
		"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.StringVar(&hiveProbeAddr, "hive-health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&hiveDataProbeAddr, "hive-data-health-probe-bind-address", ":8082", "The address the probe endpoint binds to.")
	flag.BoolVar(&secureMetrics, "metrics-secure", true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	kernelIDBytes, err := os.ReadFile(hiveDiscover.KernelIDPath)
	if err != nil {
		setupLog.Error(err, "Cannot read kerrnel boot ID at"+hiveDiscover.KernelIDPath)
		os.Exit(1)
	}
	hiveDiscover.KernelID = string(kernelIDBytes)
	hiveDiscover.KernelID = strings.TrimSpace(hiveDiscover.KernelID)

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOpts,
	})

	// Metrics endpoint is enabled in 'config/default/kustomization.yaml'. The Metrics options configure the server.
	// More info:
	// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/metrics/server
	// - https://book.kubebuilder.io/reference/metrics.html
	metricsServerOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
		// TODO(user): TLSOpts is used to allow configuring the TLS config used for the server. If certificates are
		// not provided, self-signed certificates will be generated by default. This option is not recommended for
		// production environments as self-signed certificates do not offer the same level of trust and security
		// as certificates issued by a trusted Certificate Authority (CA). The primary risk is potentially allowing
		// unauthorized access to sensitive metrics data. Consider replacing with CertDir, CertName, and KeyName
		// to provide certificates, ensuring the server communicates using trusted and secure certificates.
		TLSOpts: tlsOpts,
	}

	if secureMetrics {
		// FilterProvider is used to protect the metrics endpoint with authn/authz.
		// These configurations ensure that only authorized users and service accounts
		// can access the metrics endpoint. The RBAC are configured in 'config/rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/metrics/filters#WithAuthenticationAndAuthorization
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	// HiveData manager
	hiveDataMgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: hiveDataProbeAddr,
		LeaderElection:         true,
		LeaderElectionID:       hiveDiscover.KernelID,
	})
	if err != nil {
		setupLog.Error(err, "unable to start hiveData manager")
		os.Exit(1)
	}

	// Hive manager
	hiveMgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: hiveProbeAddr,
	})
	if err != nil {
		setupLog.Error(err, "unable to start hive manager")
		os.Exit(1)
	}

	if err = (&controller.HiveReconciler{
		Client: hiveMgr.GetClient(),
		Scheme: hiveMgr.GetScheme(),
	}).SetupWithManager(hiveMgr); err != nil {
		setupLog.Error(err, "unable to create hive controller", "controller", "Hive")
		os.Exit(1)
	}

	if err = (&controller.HiveDataReconciler{
		Client: hiveDataMgr.GetClient(),
		Scheme: hiveDataMgr.GetScheme(),
	}).SetupWithManager(hiveDataMgr); err != nil {
		setupLog.Error(err, "unable to create hiveData controller", "controller", "HiveData")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := hiveMgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := hiveMgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	if err := hiveDataMgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := hiveDataMgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting hive managers")

	go func() {
		if err := hiveMgr.Start(context.Background()); err != nil {
			setupLog.Error(err, "problem running hive manager")
			os.Exit(1)
		}
	}()

	if err := hiveDataMgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running hiveData manager")
		os.Exit(1)
	}

	// Cleanup
	if hiveDiscover.ContainerdClient != nil {
		hiveDiscover.ContainerdClient.Close()
	}
	if hiveDiscover.RingbuffReader != nil {
		hiveDiscover.RingbuffReader.Close()
	}
	hiveDiscover.Objs.Close()
	hiveDiscover.KeyProbe.Close()
}
