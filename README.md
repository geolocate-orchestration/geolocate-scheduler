# aida-scheduler

[![Test](https://github.com/aida-dos/aida-scheduler/actions/workflows/test.yml/badge.svg?branch=develop)](https://github.com/aida-dos/aida-scheduler/actions/workflows/test.yml)

IoT and other data-intensive systems produce enormous amounts of data to be processed. In this project, we explore the
Edge Computing concept as additional computational power to the Cloud nodes, allowing data process computations to occur
closer to their source.

Using Kubernetes/KubeEdge to manage the data processing workloads, the aida-scheduler aims towards substantial
improvements in scalability levels, reducing the request latency and network usage, by scheduling those workloads in
Edge Nodes based on the geographical location of the data and resource availability.

## Usage

> :warning: First, make sure you have deployed or are running the aida-controller in the cluster by following the instructions in
the [project repository](https://github.com/aida-dos/aida-controller). This controller is responsible for the management and reconciliation of the AIDA EdgeDeployments, which are the
the resource type of our workloads.

### Configuration - Environment variables

`ALGORITHM` defines the method used to select target nodes for workload placement available algorithms are:
* `location`(default): where nodes are filtered based on if they have enough available resources, and
  the target node is selected based on the deployment required/preferred locations,
* `naivelocation`: where the target node is selected based on the deployment required/preferred locations
* `random`: where any edge node is a good fit for the given workloads


### Deployment

```shell
kubectl apply -f examples/release_scheduler_crd.yaml
```

#### Build custom docker image
```shell
docker build -t aida-scheduler .
```

### Development

#### Run the aida-scheduler

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
    ALGORITHM=naivelocation go run main.go
    ```

#### Lint
```shell
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.39.0
golangci-lint ./...
```

#### Testing and Coverage
```shell
go test --coverprofile=coverage.out ./...
go tool cover -html=coverage.out 
```

#### Format

```shell
go fmt ./...
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
