# scheduler

### Kubernetes Pod Lifecycle

1. A pod is created and its desired state is saved to `etcd` with the node name unfilled
2. The scheduler somehow notices that there is a new pod with no node bound
3. It finds the node that best fits that pod
4. Tells the `apiserver` to `bind` the pod to the node -> saves the new desired state to `etcd`
5. `Kubelets` are watching bound pods through the `apiserver`, and start the containers on the particular node

### Scheduler Lifecycle

1. A loop to watch the unbound pods in the cluster through querying the `apiserver`
2. Some custom logic that finds the best node for a pod
3. A request to the bind endpoint on the `apiserver`

## Development

```
ksync watch &
ksync create --selector=component=aida-scheduler --reload=false --local-read-only=true $(pwd) /code
```
