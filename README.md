# <name-tbd>

This tool automatically restarts Pods when the image they use is rebuilt or
updated.

It's especially useful for [Minikube] and it can be deployed to Minikube as a
Pod.

[Minikube]: https://github.com/kubernetes/minikube

## Installing

Currently you can only build from the source code:

```sh
minikube start
eval $(minikube docker-env) .
docker build -t killa .
kubectl apply -f ./yaml/single-pod.yaml
```

It should be running:

```
$ kubectl get pods -n kube-system
NAME                          READY     STATUS    RESTARTS   AGE
killa                         1/1       Running   0          20s
```

```
$ kubectl logs -n kube-system -f killa
2017/11/29 22:33:26 connected kubernetes apiserver (v1.8.0)
2017/11/29 22:33:26 connected docker api (api: v1.30, version: 17.06.0-ce)
2017/11/29 22:33:26 [TRACK] pod default/hello-5766f88f9c-d5rqf
2017/11/29 22:33:26 [TRACK] pod doctor/nginx-7cbc4b4d9c-kh244
2017/11/29 22:33:26 [TRACK] pod default/hello-5766f88f9c-67lgz
```

## See in action

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
