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
	"os"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// isInCluster is the same heuristic used by rest.InClusterConfig() to
// determine whether it's an in-cluster execution environment.
func isInCluster() bool {
	return os.Getenv("KUBERNETES_SERVICE_HOST") != "" &&
		os.Getenv("KUBERNETES_SERVICE_PORT") != ""
}

// kubernetesClient loads a Kubernetes client using in-cluster configuration if
// it detects it's running inside the cluster. Otherwise it uses the default
// loading rules, such as the well-known path and the environment variable.
func kubernetesClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	if isInCluster() {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, errors.Wrap(err, "failed to load in-cluster config")
		}
	} else {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules,
			&clientcmd.ConfigOverrides{})
		config, err = kubeConfig.ClientConfig()
		if err != nil {
			return nil, errors.Wrap(err, "failed to load the kube config")
		}
	}

	clients, err := kubernetes.NewForConfig(config)
	return clients, errors.Wrap(err, "cannot initialize a kubernetes client with loaded configuration")
}
