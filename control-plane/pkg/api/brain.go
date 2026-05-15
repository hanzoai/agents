// Copyright 2026 Hanzo AI.
// Licensed under the Apache License, Version 2.0.

package api

import (
	"errors"
	"net/http"

	hanzov1 "github.com/hanzoai/operator/api/v1alpha1"
	"github.com/hanzoai/iam-sdk/go/tenant"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// brainName is fixed per-org: one Brain CR per org, named "brain" in the
// org's namespace. Simpler invariant than letting customers pick a name.
const brainName = "brain"

// BrainPut is the PUT request body. All fields optional; absent fields
// preserve current spec.
type BrainPut struct {
	Image    *string `json:"image,omitempty"`
	Replicas *int32  `json:"replicas,omitempty"`
	Storage  *string `json:"storage,omitempty"`
	Host     *string `json:"host,omitempty"`
}

// BrainView is the GET response.
type BrainView struct {
	Name      string             `json:"name"`
	Namespace string             `json:"namespace"`
	Spec      hanzov1.BrainSpec  `json:"spec"`
	Status    hanzov1.BrainStatus `json:"status"`
}

func (s *Server) handleGetBrain(w http.ResponseWriter, r *http.Request, id tenant.Identity) {
	ns := namespaceFor(id.Org)
	var b hanzov1.Brain
	key := types.NamespacedName{Namespace: ns, Name: brainName}
	if err := s.Client.Get(r.Context(), key, &b); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "brain not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, BrainView{
		Name: b.Name, Namespace: b.Namespace, Spec: b.Spec, Status: b.Status,
	})
}

func (s *Server) handlePutBrain(w http.ResponseWriter, r *http.Request, id tenant.Identity) {
	var put BrainPut
	if !readJSON(w, r, &put) {
		return
	}
	ns := namespaceFor(id.Org)

	var b hanzov1.Brain
	key := types.NamespacedName{Namespace: ns, Name: brainName}
	err := s.Client.Get(r.Context(), key, &b)
	create := errors.Is(err, nil) == false && apierrors.IsNotFound(err)
	if err != nil && !apierrors.IsNotFound(err) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if create {
		b = hanzov1.Brain{
			ObjectMeta: metav1.ObjectMeta{Name: brainName, Namespace: ns},
			Spec:       defaultBrainSpec(),
		}
		applyBrainPut(&b.Spec, put)
		if err := s.Client.Create(r.Context(), &b); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, BrainView{
			Name: b.Name, Namespace: b.Namespace, Spec: b.Spec, Status: b.Status,
		})
		return
	}

	applyBrainPut(&b.Spec, put)
	if err := s.Client.Update(r.Context(), &b); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, BrainView{
		Name: b.Name, Namespace: b.Namespace, Spec: b.Spec, Status: b.Status,
	})
}

func (s *Server) handleDeleteBrain(w http.ResponseWriter, r *http.Request, id tenant.Identity) {
	ns := namespaceFor(id.Org)
	b := &hanzov1.Brain{ObjectMeta: metav1.ObjectMeta{Name: brainName, Namespace: ns}}
	if err := s.Client.Delete(r.Context(), b); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "brain not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func defaultBrainSpec() hanzov1.BrainSpec {
	one := int32(1)
	return hanzov1.BrainSpec{
		Image: hanzov1.ImageSpec{
			Repository: "ghcr.io/hanzoai/brain",
			Tag:        "latest",
		},
		Replicas: &one,
		Storage:  hanzov1.BrainStorageSpec{Size: "10Gi"},
	}
}

func applyBrainPut(spec *hanzov1.BrainSpec, put BrainPut) {
	if put.Image != nil && *put.Image != "" {
		repo, tag := splitImage(*put.Image)
		spec.Image.Repository = repo
		if tag != "" {
			spec.Image.Tag = tag
		}
	}
	if put.Replicas != nil {
		v := *put.Replicas
		spec.Replicas = &v
	}
	if put.Storage != nil && *put.Storage != "" {
		spec.Storage.Size = *put.Storage
	}
	if put.Host != nil {
		spec.Host = *put.Host
	}
}

// splitImage breaks "repo:tag" into ("repo", "tag"). When no tag is
// present returns (image, "").
func splitImage(image string) (string, string) {
	for i := len(image) - 1; i >= 0; i-- {
		switch image[i] {
		case ':':
			return image[:i], image[i+1:]
		case '/':
			return image, ""
		}
	}
	return image, ""
}

// brainClient is a tiny adapter used by tests to assert calls; kept
// here so the file is the unit of inspection.
var _ ctrlclient.Client = (ctrlclient.Client)(nil)
