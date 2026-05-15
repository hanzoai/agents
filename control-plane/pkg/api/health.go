// Copyright 2026 Hanzo AI.
// Licensed under the Apache License, Version 2.0.

package api

import (
	"net/http"

	corev1 "k8s.io/api/core/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// handleHealth always returns 200; the process is alive.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleReady probes the K8s API by listing namespaces (limit=1). A
// successful list proves we have an authenticated, working client.
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	var nsList corev1.NamespaceList
	if err := s.Client.List(r.Context(), &nsList, ctrlclient.Limit(1)); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "not-ready", "error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
