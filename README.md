# `podfresh`

podfresh automatically restarts Pods when the image they use is rebuilt or
updated. It’s suitable for single-instance Kubernetes clusters, such as
[Minikube].

It works by listening to docker-engine image tag events, and deletes the Pods
running an updated image. It assumes that the deleted Pods will be replaced with
new ones. Therefore you should use a high-level controller such as [Deployment],
and not use Pods directly in your manifests.

[Minikube]: https://github.com/kubernetes/minikube
[Deployment]: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/

## Try it out

Deploy to your Minikube cluster with [provided manifest file](yaml/test-deployment.yaml):

```sh
minikube start
kubectl apply -f ./yaml/test-deployment.yaml
```

Check the Pod is running:

```
$ kubectl get pods -n kube-system
NAME                          READY     STATUS    RESTARTS   AGE
killa-5cbb9955cb-tcvmb        1/1       Running   0          20s
```

Verify it connected to Docker and Kubernetes APIs.
```
$ kubectl logs -n kube-system killa-5cbb9955cb-tcvmb
2017/11/29 22:33:26 connected kubernetes apiserver (v1.8.0)
2017/11/29 22:33:26 connected docker api (api: v1.30, version: 17.06.0-ce)
2017/11/29 22:33:26 [TRACK] pod default/hello-5766f88f9c-d5rqf
2017/11/29 22:33:26 [TRACK] pod default/hello-5766f88f9c-67lgz
...
```

## See it in action

Get some test images:

```
eval $(minikube docker-env)
docker pull gcr.io/google-samples/hello-app:1.0
docker pull gcr.io/google-samples/hello-app:2.0
docker tag  gcr.io/google-samples/hello-app:1.0 hello
```

Run a 2-replica Deployment and NodePort Service with `hello` image:

```
kubectl run hello --image=gcr.io/google-samples/hello-app:1.0 \
  --replicas=2 --port 8080

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

Re-tag the `hello` image with `2.0`:

```
docker tag gcr.io/google-samples/hello-app:2.0 hello
```

Visit the app again (note the `2.0.0`):

```
$ curl "$URL"
Hello, world!
Version: 2.0.0
Hostname: hello-5766f88f9c-h88df
```

#### Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for more information.

-----

This is not an official Google product.
