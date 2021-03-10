# scheduler

### What's next

Using only informers for caching is still not enough for a production grade scheduler,
as it needs to track the current resource requests on nodes, pods running on each of the nodes,
along a few other things. The default scheduler contains a custom cache implementation on top of
the informers to achieve this. And there are also other things like handling errors during scheduling,
using a proper queue implementation, achieving high availability, and of course writing some meaningful
business logic to find a proper node just to name a few.

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
ksync watch

# If config not created:
ksync create --selector=component=aida-scheduler --reload=false --local-read-only=true $(pwd) /code

keti aida-scheduler-55dd5b4747-l6l6z -- sh
  cd /code
  go run main.go
```
