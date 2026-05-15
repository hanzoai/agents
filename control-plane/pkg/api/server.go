// Copyright 2026 Hanzo AI.
// Licensed under the Apache License, Version 2.0.

// Package api is the HTTP front door for the Hanzo Agents control plane.
//
// Three decomplected concerns, composed in Server:
//
//  1. authn — delegate to tenant.Require (iam-sdk).
//  2. CR write — delegate to controller-runtime client.
//  3. status read — delegate to the same client.
//
// No business logic. The operator reconciles Brain and Bot CRs into K8s.
package api

import (
	"encoding/json"
	"net/http"

	hanzov1 "github.com/hanzoai/operator/api/v1alpha1"
	"github.com/hanzoai/iam-sdk/go/tenant"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Scheme registers core/v1 + hanzo.ai/v1alpha1 for the controller-runtime client.
// Exported so callers (production main, tests with fake client) can share it.
var Scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(Scheme))
	utilruntime.Must(hanzov1.AddToScheme(Scheme))
}

// PodLogger streams pod logs for a deployment. Decoupled from the CR
// client because controller-runtime's client doesn't expose log
// streaming — that's a clientset subresource. Set to nil to disable.
type PodLogger interface {
	TailLogs(w http.ResponseWriter, r *http.Request, namespace, deployment string) error
}

// Server is the composition root. Single-noun.
type Server struct {
	// Client is the controller-runtime client used to read and write CRs.
	Client ctrlclient.Client
	// Verifier authenticates incoming requests.
	Verifier tenant.Verifier
	// Logger optionally streams pod logs. May be nil.
	Logger PodLogger
}

// Handler returns an http.Handler with all /v1 routes mounted. The
// returned handler enforces tenant.Require on every /v1 endpoint;
// /health and /ready are public.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Public.
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /ready", s.handleReady)

	// Brain.
	mux.HandleFunc("GET /v1/orgs/{org}/brain", s.requireOrg(s.handleGetBrain))
	mux.HandleFunc("PUT /v1/orgs/{org}/brain", s.requireOrg(s.handlePutBrain))
	mux.HandleFunc("DELETE /v1/orgs/{org}/brain", s.requireOrg(s.handleDeleteBrain))

	// Bot.
	mux.HandleFunc("GET /v1/orgs/{org}/bots", s.requireOrg(s.handleListBots))
	mux.HandleFunc("POST /v1/orgs/{org}/bots", s.requireOrg(s.handleCreateBot))
	mux.HandleFunc("GET /v1/orgs/{org}/bots/{name}", s.requireOrg(s.handleGetBot))
	mux.HandleFunc("DELETE /v1/orgs/{org}/bots/{name}", s.requireOrg(s.handleDeleteBot))
	mux.HandleFunc("POST /v1/orgs/{org}/bots/{name}/start", s.requireOrg(s.handleStartBot))
	mux.HandleFunc("POST /v1/orgs/{org}/bots/{name}/stop", s.requireOrg(s.handleStopBot))
	mux.HandleFunc("GET /v1/orgs/{org}/bots/{name}/logs", s.requireOrg(s.handleBotLogs))

	return tenant.Require(s.Verifier, mux)
}

// requireOrg wraps a handler so that the {org} path parameter must
// match the Identity.Org placed on the context by tenant.Require. Any
// mismatch produces 403.
func (s *Server) requireOrg(fn func(http.ResponseWriter, *http.Request, tenant.Identity)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := tenant.FromContext(r.Context())
		if !ok {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		if r.PathValue("org") != id.Org {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		fn(w, r, id)
	}
}

// namespaceFor derives the K8s namespace from the org slug. One org,
// one namespace; the operator scopes all CRs to it.
func namespaceFor(org string) string { return "org-" + org }

// writeJSON writes obj as JSON with HTTP status code.
func writeJSON(w http.ResponseWriter, code int, obj any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(obj)
}

// readJSON decodes the request body into obj. Returns 400 on failure.
func readJSON(w http.ResponseWriter, r *http.Request, obj any) bool {
	if r.Body == nil {
		http.Error(w, "empty body", http.StatusBadRequest)
		return false
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(obj); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

// asPodList is a small helper used in tests; placed here to keep the
// concrete corev1 import alive without leaking into handlers.
var _ = corev1.PodList{}
