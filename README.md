# raft
Go implementation of [Raft Consensus Protocol](https://raft.github.io/).

## Run Raft locally in minikube

### Prerequisite
1. docker [installed](https://docs.docker.com/get-docker/).
2. minikube [installed](https://minikube.sigs.k8s.io/docs/start/).
3. kubectl [installed](https://kubernetes.io/docs/tasks/tools/).

### Steps

1. Start minikube
```bash
$ minikube start
```

2. Build raft image locally using the Dockerfile
```bash
$ make docker-build tag=0.0.1 
```

3. Apply the Kubernetes manifests for the Raft cluster
```bash
# It includes 3 nodes right now.
$ kubectl apply -f demo/ --context minikube 
```

4. Verify the raft pod is bootstrapped and able to serve traffic
```bash
$ kubectl -n raft get pods --context minikube
NAME                    READY   STATUS    RESTARTS   AGE
raft-84cdb86794-chncb   1/1     Running   0          14m 
$ kubectl exec -it $ECHO_CLIENT_POD -n echo-client -- \
 env bash curl -I raft.raft.svc.cluster.local:8080/ping
HTTP/1.1 200 OK
Server: fasthttp
Date: Tue, 21 Jun 2022 06:02:26 GMT
Content-Type: application/json
```
