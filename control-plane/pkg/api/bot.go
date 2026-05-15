// Copyright 2026 Hanzo AI.
// Licensed under the Apache License, Version 2.0.

package api

import (
	"net/http"
	"regexp"

	hanzov1 "github.com/hanzoai/operator/api/v1alpha1"
	"github.com/hanzoai/iam-sdk/go/tenant"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// botNameRE enforces DNS-1123 label rules so the name maps cleanly to
// the underlying Deployment + Service.
var botNameRE = regexp.MustCompile(`^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$`)

// BotCreate is the POST body.
type BotCreate struct {
	Name     string             `json:"name"`
	Channels []hanzov1.BotChannel `json:"channels,omitempty"`
	Image    string             `json:"image,omitempty"`
}

// BotView is the GET / list response shape.
type BotView struct {
	Name      string             `json:"name"`
	Namespace string             `json:"namespace"`
	Spec      hanzov1.BotSpec    `json:"spec"`
	Status    hanzov1.BotStatus  `json:"status"`
}

func (s *Server) handleListBots(w http.ResponseWriter, r *http.Request, id tenant.Identity) {
	ns := namespaceFor(id.Org)
	var list hanzov1.BotList
	if err := s.Client.List(r.Context(), &list, ctrlclient.InNamespace(ns)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := make([]BotView, 0, len(list.Items))
	for i := range list.Items {
		b := &list.Items[i]
		out = append(out, BotView{Name: b.Name, Namespace: b.Namespace, Spec: b.Spec, Status: b.Status})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (s *Server) handleCreateBot(w http.ResponseWriter, r *http.Request, id tenant.Identity) {
	var body BotCreate
	if !readJSON(w, r, &body) {
		return
	}
	if !botNameRE.MatchString(body.Name) {
		http.Error(w, "invalid bot name (DNS-1123 label required)", http.StatusBadRequest)
		return
	}
	ns := namespaceFor(id.Org)

	one := int32(1)
	bot := hanzov1.Bot{
		ObjectMeta: metav1.ObjectMeta{Name: body.Name, Namespace: ns},
		Spec: hanzov1.BotSpec{
			Brain:    hanzov1.BotBrainRef{Name: brainName},
			Image:    defaultBotImage(body.Image),
			Channels: body.Channels,
			Replicas: &one,
		},
	}
	if err := s.Client.Create(r.Context(), &bot); err != nil {
		if apierrors.IsAlreadyExists(err) {
			http.Error(w, "bot already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, BotView{Name: bot.Name, Namespace: bot.Namespace, Spec: bot.Spec, Status: bot.Status})
}

func (s *Server) handleGetBot(w http.ResponseWriter, r *http.Request, id tenant.Identity) {
	bot, ok := s.loadBot(w, r, id)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, BotView{Name: bot.Name, Namespace: bot.Namespace, Spec: bot.Spec, Status: bot.Status})
}

func (s *Server) handleDeleteBot(w http.ResponseWriter, r *http.Request, id tenant.Identity) {
	ns := namespaceFor(id.Org)
	name := r.PathValue("name")
	bot := &hanzov1.Bot{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns}}
	if err := s.Client.Delete(r.Context(), bot); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "bot not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleStartBot(w http.ResponseWriter, r *http.Request, id tenant.Identity) {
	s.setReplicas(w, r, id, 1)
}

func (s *Server) handleStopBot(w http.ResponseWriter, r *http.Request, id tenant.Identity) {
	s.setReplicas(w, r, id, 0)
}

func (s *Server) handleBotLogs(w http.ResponseWriter, r *http.Request, id tenant.Identity) {
	if s.Logger == nil {
		http.Error(w, "log streaming not enabled", http.StatusServiceUnavailable)
		return
	}
	ns := namespaceFor(id.Org)
	name := r.PathValue("name")
	if err := s.Logger.TailLogs(w, r, ns, name); err != nil {
		// Logger may have already written headers; only emit if not yet sent.
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) setReplicas(w http.ResponseWriter, r *http.Request, id tenant.Identity, n int32) {
	bot, ok := s.loadBot(w, r, id)
	if !ok {
		return
	}
	v := n
	bot.Spec.Replicas = &v
	if err := s.Client.Update(r.Context(), bot); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, BotView{Name: bot.Name, Namespace: bot.Namespace, Spec: bot.Spec, Status: bot.Status})
}

func (s *Server) loadBot(w http.ResponseWriter, r *http.Request, id tenant.Identity) (*hanzov1.Bot, bool) {
	ns := namespaceFor(id.Org)
	name := r.PathValue("name")
	var bot hanzov1.Bot
	if err := s.Client.Get(r.Context(), types.NamespacedName{Namespace: ns, Name: name}, &bot); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "bot not found", http.StatusNotFound)
			return nil, false
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, false
	}
	return &bot, true
}

func defaultBotImage(image string) hanzov1.ImageSpec {
	if image == "" {
		return hanzov1.ImageSpec{Repository: "ghcr.io/hanzoai/bot", Tag: "latest"}
	}
	repo, tag := splitImage(image)
	if tag == "" {
		tag = "latest"
	}
	return hanzov1.ImageSpec{Repository: repo, Tag: tag}
}
