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
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/tools/cache"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-signalChan
		log.Printf("received signal: %s", sig.String())
		cancel()
	}()

	tagEvent := filters.NewArgs()
	tagEvent.Add("type", "image")
	tagEvent.Add("event", "tag")

	k8s, err := kubernetesClient()
	if err != nil {
		log.Fatal(err)
	}
	k8sv, err := k8s.ServerVersion()
	if err != nil {
		log.Fatal("failed to connect to kubernetes")
	}
	log.Printf("connected kubernetes apiserver (%s)", k8sv.GitVersion)

	d, err := dockerclient.NewEnvClient()
	if err != nil {
		log.Fatal(errors.Wrap(err, "cannot create docker client"))
	}
	d.NegotiateAPIVersion(ctx)
	dv, err := d.ServerVersion(ctx)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to connect to docker api"))
	}
	log.Printf("connected docker api (api: v%s, version: %s)", dv.APIVersion, dv.Version)

	podHandler := &podDeletionHandler{pods: newRegistry()}
	tagCh := podHandler.Start(ctx, k8s.CoreV1())

	podWatcher := podWatchController(k8s, podHandler)
	go podWatcher.Run(ctx.Done())

	ch, errCh := d.Events(ctx, types.EventsOptions{Filters: tagEvent})
	for {
		select {
		case err := <-errCh:
			select {
			case <-ctx.Done():
				log.Println("stopping event listener due to cancellation")
				os.Exit(0)
			default:
				panic(err) // TODO(ahmetb) handle gracefully
			}

		case e := <-ch:
			// tag will be in format IMAGE:TAG or IMAGE:latest as it comes
			// from the Docker API (v1.32 at the time of writing).
			tag := e.Actor.Attributes["name"]
			tagCh <- tag
		case <-ctx.Done():
			break
		}
	}
}

func podWatchController(k8s *kubernetes.Clientset, pods *podDeletionHandler) cache.Controller {
	restClient := k8s.CoreV1().RESTClient()
	lw := cache.NewListWatchFromClient(restClient, "pods", corev1.NamespaceAll, fields.Everything())
	_, controller := cache.NewInformer(lw,
		&corev1.Pod{},
		time.Second*5,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod, ok := obj.(*corev1.Pod)
				if !ok {
					log.Fatalf("list/watch returned non-pod object: %T", obj)
				}
				pods.Track(pod)
			},
			DeleteFunc: func(obj interface{}) {
				pod, ok := obj.(*corev1.Pod)
				if !ok {
					log.Fatalf("list/watch returned non-pod object: %T", obj)
				}
				pods.Untrack(pod)
			},
		},
	)
	return controller
}
