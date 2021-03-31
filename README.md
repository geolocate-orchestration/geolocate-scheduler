# aida-scheduler

[![Test](https://github.com/aida-dos/aida-scheduler/actions/workflows/test.yml/badge.svg?branch=develop)](https://github.com/aida-dos/aida-scheduler/actions/workflows/test.yml)

## Deployment

```shell
kubectl apply -f examples/release_scheduler_crd.yaml
```

### Build custom docker image
```shell
docker build -t aida-scheduler .
```

## Development

### Run the aida-controller

First, make sure you have deployed or are running the aida-controller in the cluster by following the instructions in
the [project repository](https://github.com/aida-dos/aida-controller).

This controller is responsible for the management and reconciliation of the AIDA EdgeDeployments, which are the
the resource type of our workloads.

### Run the aida-scheduler

To develop aida-scheduler in a Kubernetes cluster we use [ksync](https://github.com/ksync/ksync)
to sync files between our local system, and the cluster.

1. Install ksync. You can follow ksync installation steps [here](https://github.com/ksync/ksync#installation).

2. Create a deployment where we will run the scheduler by applying the
   [example/dev_scheduler_crd.yaml](example/dev_scheduler_crd.yaml).
   ```shell
    kubectl apply -f example/dev_scheduler_crd.yaml
    ```
   
3. If not done before then create a ksync configuration for the current folder.
    ```shell
    ksync create --selector=component=aida-scheduler --reload=false --local-read-only=true $(pwd) /code
    ```

4. Start ksync update system
    ```shell
    ksync watch
    ```

5. Run the scheduler in the cluster pod
    ```shell
    kubectl exec -it $(kubectl get pod -n kube-system | grep aida-scheduler | awk '{print $1}') -- sh
    cd /code
    go run main.go
    ```

### Manage nodes

The aida-scheduler only manages Edge nodes, because the main purpose is to allow application workload to be deployed
near the source of data to be processed. Therefore, the node controller filters nodes by 'node-role.kubernetes.io/edge'
labels.

- To add the label
    ```shell
    kubectl label node node0 --overwrite node-role.kubernetes.io/edge= 
    ```

- To remove the label
    ```shell
    kubectl label node node0 --overwrite node-role.kubernetes.io/edge-
    ```
  
### Deploy workloads

Apply any of the workload [examples](examples)

- No location set
    ```shell
    kubectl apply -f examples/workload_no_set_location.yaml
    ```

- Required location
    ```shell
    kubectl apply -f examples/workload_required_location.yaml
    ```

- Preferred location
    ```shell
    kubectl apply -f examples/workload_preferred_location.yaml
    ```

### Lint
```shell
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.39.0
golangci-lint ./...
```

### Testing and Coverage
```shell
go test --coverprofile=coverage.out ./...
go tool cover -html=coverage.out 
```

### Format

```shell
go fmt ./...
```
