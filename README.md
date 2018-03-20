# freshpod [![](https://img.shields.io/badge/THANKS-md-ff69b4.svg)](THANKS.md)

freshpod helps you automatically restart containers on Kubernetes when their image
is updated.

It is designed for single-instance Kubernetes clusters, such as [Minikube] or
[Docker for Windows/Mac][dfm].

freshpod detects you rebuilt an image and it deletes the Kubernetes Pods are
running that image. This way, your workload controller (such as [Deployment])
will create new Pods running the new image.

> :new: **Check out [Skaffold]**, a new tool by Google that simplifies local Kubernetes
> development experience. Skaffold supersedes freshpod.

[Minikube]: https://github.com/kubernetes/minikube
[dfm]: https://docs.docker.com/docker-for-mac/kubernetes/
[Pods]: https://kubernetes.io/docs/concepts/workloads/pods/pod/
[Deployment]: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/
[Skaffold]: https://github.com/GoogleCloudPlatform/skaffold

## Demo

This demo shows refactoring an application and running `docker build`
will automatically restart the application with the new image on [Minikube]:

[![A command line demo of freshpod replacing pods when the image is updated](img/freshpod-demo.gif)](https://asciinema.org/a/dD9UhCIaPw13znirhmGUnNJtd)

## Install on Minikube

freshpod is available as an add-on for [Minikube]. You just need to enable it:

    minikube addons enable freshpod

## Install on “Docker for Mac/Windows”

If you’re using [Kubernertes on Docker for Mac/Windows][dfm], you can directly
apply [the manifest](https://github.com/kubernetes/minikube/blob/master/deploy/addons/freshpod/freshpod-rc.yaml)
used by Minikube:

    kubectl apply -f https://raw.githubusercontent.com/kubernetes/minikube/ec1b443722227428bd2b23967e1b48d94350a5ac/deploy/addons/freshpod/freshpod-rc.yaml

## Try it out!

Get some test images and tag the `:1.0` image as `hello:latest`:

```sh
eval $(minikube docker-env) # not necessary for docker-for-mac/windows
docker pull gcr.io/google-samples/hello-app:1.0
docker pull gcr.io/google-samples/hello-app:2.0
docker tag  gcr.io/google-samples/hello-app:1.0 hello:latest
```

Run a 2-replica Deployment and NodePort Service with `hello:latest` image:

```sh
kubectl run hello --image=hello --port 8080 --replicas=2 \
  --image-pull-policy=IfNotPresent

kubectl expose deploy/hello --type=NodePort
```

Visit the app (note the `1.0.0`):

```
$ URL=$(minikube service hello --url)
$ curl "$URL"
Hello, world!
Version: 1.0.0
Hostname: hello-5766f88f9c-67lgz
```

Re-tag the `hello:latest` image with the `2.0` version:

```sh
docker tag gcr.io/google-samples/hello-app:2.0 hello
```

Visit the app again (note the version has changed to `2.0.0`):

```
$ curl "$URL"
Hello, world!
Version: 2.0.0
Hostname: hello-5766f88f9c-h88df
```

-----

#### Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for more information.

This is not an official Google product.
