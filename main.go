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

const (
	dockerSock = "/var/run/docker.sock"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	os.Setenv("DOCKER_HOST", "unix://"+dockerSock)

	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		<-signalChan
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
	dv, err := d.ServerVersion(ctx)
	if err != nil {
		log.Fatal("failed to connect to docker api")
	}
	log.Printf("connected docker api (api: v%s, version: %s)", dv.APIVersion, dv.Version)

	podWatcher := podWatchController(k8s)
	go podWatcher.Run(ctx.Done())

	ch, errCh := d.Events(ctx, types.EventsOptions{Filters: tagEvent})
	for {
		select {
		case err := <-errCh:
			panic(err) // TODO(ahmetb) handle gracefully
		case e := <-ch:
			tag := e.Actor.Attributes["name"]
			go func() {
				if err := handleTag(tag); err != nil {
					log.Println(err)
				}
			}()
		case <-ctx.Done():
			break
		}
	}
}

func podWatchController(k8s *kubernetes.Clientset) cache.Controller {
	restClient := k8s.CoreV1().RESTClient()
	lw := cache.NewListWatchFromClient(restClient, "pods", corev1.NamespaceAll, fields.Everything())
	_, controller := cache.NewInformer(lw,
		&corev1.Pod{},
		time.Second*5,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod, ok := obj.(*corev1.Pod)
				if !ok {
					log.Fatalf("list/watch returned non-pod object: %T", pod)
				}
				if err := trackPod(pod); err != nil {
					// TODO(ahmetb) if workqueue used here maybe we can
					// requeue somehow.
					log.Println(errors.Wrap(err, "could not track pod"))
				}
			},
			DeleteFunc: func(obj interface{}) {
				pod, ok := obj.(*corev1.Pod)
				if !ok {
					log.Fatalf("list/watch returned non-pod object: %T", pod)
				}
				if err := untrackPod(pod); err != nil {
					// TODO(ahmetb) if workqueue used here maybe we can
					// requeue somehow.
					log.Println(errors.Wrap(err, "could not untrack pod"))
				}
			},
		},
	)
	return controller
}

func handleTag(tag string) error {
	// TODO(ahmetb) implement
	return nil
}

func trackPod(pod *corev1.Pod) error {
	// TODO(ahmetb) implement
	log.Printf("handling pod %s", pod.GetName())
	return nil
}

func untrackPod(pod *corev1.Pod) error {
	// TODO(ahmetb) implement
	log.Printf("forgetting pod %s", pod.GetName())
	return nil
}
