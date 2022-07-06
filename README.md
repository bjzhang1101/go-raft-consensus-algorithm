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

2. Setup Docker build env
```bash
$ eval $(minikube docker-env)
```

3. Build raft image locally using the Dockerfile
```bash
$ make docker-build tag=0.0.1 
```

4. Apply the Kubernetes manifests for the Raft cluster
```bash
# It includes 3 nodes right now.
$ kubectl apply -f demo/ 
```

5. Verify the raft pod is bootstrapped and able to serve traffic
```bash
$ kubectl -n raft get pods --context minikube
NAME                       READY   STATUS    RESTARTS   AGE
raft-01-78cc5d54fc-mtczj   1/1     Running   0          21m
raft-02-7dfc8744c-n9sqk    1/1     Running   0          21m
raft-03-86bcd95b5-msg4b    1/1     Running   0          21m

$ kubectl -n raft-client exec -it raft-client-7bc9d7884c-b4bxt -- env bash

# Check node raft-02 state.
$ root@raft-client-7bc9d7884c-b4bxt:/# curl raft-02.raft.svc:8080/state
{"state":"follower","term":1,"leader":"raft-01.raft.svc.cluster.local"}

# Add a new key-1 from the leader raft-01.
$ root@raft-client-7bc9d7884c-b4bxt:/# curl -XPOST raft-01.raft.svc:8080/operate_data \
    -H "Content-Type: application/json" \
    -d '{"key": "key-1", "value": "value-1", "action":"add"}'

# Send write request to a follower raft-02.
$ root@raft-client-7bc9d7884c-b4bxt:/# curl -XPOST raft-02.raft.svc:8080/operate_data \
    -H "Content-Type: application/json" \
    -d '{"key": "key-2", "value": "value-2", "action":"add"}'
{"success":false,"leader":"raft-01.raft.svc.cluster.local"}

# Check data in a follower raft-02.
$ root@raft-client-7bc9d7884c-b4bxt:/# curl raft-02.raft.svc:8080/get_all_data
{"data":"map[key-1:value-1]"}

# Delete key-1 from the leader raft-01.
$ root@raft-client-7bc9d7884c-b4bxt:/# curl -XPOST raft-01.raft.svc:8080/operate_data 
    -H "Content-Type: application/json" \
    -d '{"key": "key-1", "action":"del"}'
 
# Check data in a follower raft-02.
$ root@raft-client-7bc9d7884c-b4bxt:/# curl raft-02.raft.svc:8080/get_all_data
{"data":"map[]"}
```
