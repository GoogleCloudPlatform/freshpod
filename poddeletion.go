// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	corev1typed "k8s.io/client-go/kubernetes/typed/core/v1"
)

type podDeletionHandler struct {
	pods *podRegistry

	tagCh chan string
	mu    sync.Mutex
}

// Start returns a chan where image tags can be provided for deletion of pods
// running them and starts a goroutine for deletion in the background.
func (h *podDeletionHandler) Start(ctx context.Context, k8s corev1typed.CoreV1Interface) chan<- string {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.tagCh != nil {
		panic("pod deletion handler is already started")
	}
	h.tagCh = make(chan string)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case tag := <-h.tagCh:
				log.Printf("[image_tagged] %q", tag)
				go h.deletePods(k8s, tag)
			}
		}
	}()
	return h.tagCh
}

// deletePods deletes pods running the specified tag serially.
func (h *podDeletionHandler) deletePods(k8s corev1typed.CoreV1Interface, tag string) {
	pods := h.pods.get(tag)
	if len(pods) == 0 {
		log.Printf("[noop] no pods registered with image=%s", tag)
		return
	}
	for _, p := range pods {
		log.Printf("[deleting_pod] %s/%s", p.namespace, p.name)
		if err := k8s.Pods(p.namespace).Delete(p.name, nil); err != nil {
			log.Println(errors.Wrap(err, "failed to delete pod"))
		}
		log.Printf("[deleted_pod] %s/%s", p.namespace, p.name)

		// TODO(ahmetb) see if there's a better way of doing this: here we
		// unregister the pod directly, because we know we just deleted it. it's
		// faster than deletion to actually go through and come back via WATCH.
		h.pods.del(p, tag)
	}
}

// Track registers that we know the given pod exists right now.
func (h *podDeletionHandler) Track(p *corev1.Pod) {
	log.Printf("[track_pod] %s/%s", p.GetNamespace(), p.GetName())
	for _, c := range p.Spec.Containers {
		h.pods.add(pod{
			namespace: p.Namespace,
			name:      p.Name}, canonicalImage(c.Image))
	}
}

// Untrack removes the given pod from tracking list when it no longer exists.
func (h *podDeletionHandler) Untrack(p *corev1.Pod) {
	log.Printf("[untrack_pod] %s/%s", p.GetNamespace(), p.GetName())
	for _, c := range p.Spec.Containers {
		h.pods.del(pod{
			namespace: p.Namespace,
			name:      p.Name}, canonicalImage(c.Image))
	}
}

// canonicalImage adds :latest to the image tags so
func canonicalImage(img string) string {
	if !strings.Contains(img, ":") {
		// TODO(ahmetb) find better ways to add :latest. currently this detection
		// covers both "IMAGE:TAG" format and "IMAGE@sha256:DIGEST" formats.
		return fmt.Sprintf("%s:latest", img)
	}
	return img
}
