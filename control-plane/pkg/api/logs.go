// Copyright 2026 Hanzo AI.
// Licensed under the Apache License, Version 2.0.

package api

import (
	"context"
	"errors"
	"io"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// LogStreamer tails pod logs for a Bot deployment using a clientset.
// It picks the first pod matching the bot's label selector
// (app=<deployment>) and streams stdout to the response writer.
type LogStreamer struct {
	Clientset kubernetes.Interface
}

// TailLogs implements PodLogger.
func (l *LogStreamer) TailLogs(w http.ResponseWriter, r *http.Request, namespace, deployment string) error {
	if l.Clientset == nil {
		return errors.New("clientset not configured")
	}
	ctx := r.Context()
	pods, err := l.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=" + deployment,
		Limit:         1,
	})
	if err != nil {
		return err
	}
	if len(pods.Items) == 0 {
		http.Error(w, "no pods for bot", http.StatusNotFound)
		return nil
	}
	pod := pods.Items[0].Name

	follow := r.URL.Query().Get("follow") == "true"
	req := l.Clientset.CoreV1().Pods(namespace).GetLogs(pod, &corev1.PodLogOptions{
		Follow: follow,
	})
	rc, err := req.Stream(ctx)
	if err != nil {
		return err
	}
	defer rc.Close()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Pod-Name", pod)
	w.WriteHeader(http.StatusOK)
	flusher, _ := w.(http.Flusher)
	if _, err := streamWithFlush(ctx, w, rc, flusher); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

// streamWithFlush copies src to dst, flushing after each chunk so the
// client sees logs immediately when follow=true.
func streamWithFlush(ctx context.Context, dst io.Writer, src io.Reader, flusher http.Flusher) (int64, error) {
	buf := make([]byte, 4096)
	var total int64
	for {
		select {
		case <-ctx.Done():
			return total, ctx.Err()
		default:
		}
		n, err := src.Read(buf)
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return total, werr
			}
			total += int64(n)
			if flusher != nil {
				flusher.Flush()
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return total, nil
			}
			return total, err
		}
	}
}
