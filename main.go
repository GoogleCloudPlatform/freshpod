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

	"github.com/docker/docker/api/types/filters"

	"github.com/docker/docker/api/types"

	dockerclient "github.com/docker/docker/client"
)

const (
	dockerSock = "/var/run/docker.sock"
)

func main() {
	ctx := context.Background()
	os.Setenv("DOCKER_HOST", "unix://"+dockerSock)

	d, err := dockerclient.NewEnvClient()
	if err != nil {
		panic(err) // TODO(ahmetb) handle better
	}

	tagEvent := filters.NewArgs()
	tagEvent.Add("type", "image")
	tagEvent.Add("event", "tag")

	ch, errCh := d.Events(ctx, types.EventsOptions{Filters: tagEvent})
	for {
		select {
		case err := <-errCh:
			panic(err) // TODO(ahmetb) handle gracefully
		case e := <-ch:
			tag := e.Actor.Attributes["name"]
			go func() {
				if err := handle(tag); err != nil {
					log.Println(err)
				}
			}()
		}
	}
}

func handle(tag string) error {
	return nil
}
