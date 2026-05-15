// Copyright 2026 Hanzo AI.
// Licensed under the Apache License, Version 2.0.

// Binary agents serves the Hanzo Agents control plane HTTP API.
//
// It is a thin REST -> CRD facade: incoming /v1 calls are authenticated
// via the IAM gateway headers (tenant.Require) and translated into
// Brain / Bot custom resources on the management cluster. The
// hanzoai/operator reconciles those CRs into Kubernetes objects.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hanzoai/agents/control-plane/pkg/api"
	"github.com/hanzoai/iam-sdk/go/tenant"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	addr := flag.String("addr", ":6430", "HTTP listen address")
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig (empty = in-cluster)")
	auth := flag.String("auth", "iam", `auth provider: "iam" (gateway headers) or "none" (dev only)`)
	flag.Parse()

	cfg, err := buildRestConfig(*kubeconfig)
	if err != nil {
		log.Fatalf("kube config: %v", err)
	}

	cl, err := ctrlclient.New(cfg, ctrlclient.Options{Scheme: api.Scheme})
	if err != nil {
		log.Fatalf("k8s client: %v", err)
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("k8s clientset: %v", err)
	}

	v, err := buildVerifier(*auth)
	if err != nil {
		log.Fatalf("auth: %v", err)
	}

	srv := &api.Server{
		Client:   cl,
		Verifier: v,
		Logger:   &api.LogStreamer{Clientset: cs},
	}

	httpServer := &http.Server{
		Addr:              *addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		log.Printf("agents control-plane listening on %s", *addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("shutting down")
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutCancel()
	if err := httpServer.Shutdown(shutCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func buildRestConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if env := os.Getenv("KUBECONFIG"); env != "" {
		return clientcmd.BuildConfigFromFlags("", env)
	}
	return rest.InClusterConfig()
}

func buildVerifier(provider string) (tenant.Verifier, error) {
	switch provider {
	case "iam":
		return tenant.HeaderVerifier{}, nil
	case "none":
		// Dev only: trusts X-* headers without enforcing presence at the gateway.
		// Still requires non-empty X-Org-Id + X-User-Id via HeaderVerifier.
		return tenant.HeaderVerifier{}, nil
	default:
		return nil, fmt.Errorf("unknown auth provider %q", provider)
	}
}
