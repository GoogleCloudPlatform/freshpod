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

import "sync"

type pod struct{ name, namespace string }

type podRegistry struct {
	mu       sync.RWMutex
	imgToPod map[string]map[pod]struct{}
	podToImg map[pod]map[string]struct{}
}

func newRegistry() *podRegistry {
	return &podRegistry{
		imgToPod: make(map[string]map[pod]struct{}),
		podToImg: make(map[pod]map[string]struct{}),
	}
}

// add registers that pod uses the specified image. It can be called multiple
// times with different image values to register images of multiple containers
// of the same pod.
func (r *podRegistry) add(p pod, image string) {
	r.mu.Lock()

	if _, ok := r.imgToPod[image]; !ok {
		r.imgToPod[image] = make(map[pod]struct{})
	}

	r.imgToPod[image][p] = struct{}{}
	if _, ok := r.podToImg[p]; !ok {
		r.podToImg[p] = make(map[string]struct{})
	}
	r.podToImg[p][image] = struct{}{}
	r.mu.Unlock()
}

// del unregisters that the pod is using the specified image. It can be multiple
// times with different images to unregister images of multiple containers of
// the same pod.
func (r *podRegistry) del(p pod, image string) {
	r.mu.Lock()
	delete(r.imgToPod[image], p)
	if len(r.imgToPod[image]) == 0 {
		delete(r.imgToPod, image)
	}
	delete(r.podToImg[p], image)
	if len(r.podToImg[p]) == 0 {
		delete(r.podToImg, p)
	}
	r.mu.Unlock()
}

// get retrieves list of pods running the image.
func (r *podRegistry) get(image string) []pod {
	r.mu.RLock()
	out := make([]pod, len(r.imgToPod[image]))
	var i int
	for p := range r.imgToPod[image] {
		out[i] = p
		i++
	}
	r.mu.RUnlock()
	return out
}
