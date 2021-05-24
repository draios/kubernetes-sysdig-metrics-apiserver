# Kubernetes Custom Metrics Adapter for Sysdig

> :warning: This project is maintained only until September 2021. We encourage you to use [keda](https://keda.sh/)

[![Build status][1]][2]

Table of contents:

- [Kubernetes Custom Metrics Adapter for Sysdig](#kubernetes-custom-metrics-adapter-for-sysdig)
  - [Introduction](#introduction)
  - [Breaking Changes in the v2 upgrade](#breaking-changes-in-the-v2-upgrade)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Troubleshooting](#troubleshooting)
  - [Contributing](#contributing)
  - [Relevant links](#relevant-links)
  - [Credits and license](#credits-and-license)


## Introduction

`k8s-sysdig-adapter` is an implementation of the
[Custom Metrics API][custom-metrics-api-types] using the [Official Custom Metrics Adapter Server Framework][l2] and
[Sysdig Monitor][sysdig-monitor].

If you have a Kubernetes cluster and you are a Sysdig user, this adapter
enables you to create [horizontal pod autoscalers][l4] based on metrics provided
by Sysdig monitoring solution.

In the following example, you are creating an autoscaler based on the
`net.http.request.count` metric and for a `deployment` workload. The autoscaler will adjust the number of pods
deployed for our service as the metric fluctuates over or below the threshold
(`targetValue`).

```yaml
---
kind: HorizontalPodAutoscaler
apiVersion: autoscaling/v2beta1
metadata:
  name: kuard
spec:
  scaleTargetRef:
    kind: Deployment
    name: kuard
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Object
    object:
      target:
        kind: Service
        name: deployment;kuard
      metricName: net.http.request.count
      targetValue: 100
```

## Breaking Changes in the v2 upgrade
Note: The following breaking changes are applicable only for `v0.2`.
* This version requires the unified workload labels, `kubernetes.workload.name` and `kubernetes.workload.type` in your Sysdig Monitor Platform. These two workload labels have been introduced in Sysdig Monitor to check the type of workload and the workload name. In the previous version, the HPA was only asking for the namespace and the service name.

* In the previous HPA definition, the target name field had this format: `name: kuard`. In the new version, the target has different format depending on the kind of workload that you want to scale:
  *  deployment: `name: deployment;kuard`
  *  statefulset: `name: statefulset;kuard`

* Add your cluster name to your Custom Metrics API deployment as follows:
```yaml
- name: CLUSTER_NAME
  value: "YourClusterName"
```
You add a new environment variable named `CLUSTER_NAME`. Set its value to the name of the desired Kubernetes cluster in Sysdig Monitor. Ensure that you use the same name as shown in Sysdig Monitor. You can view the list of Clusters in the **Overview** tab on the Monitor UI. 

See the example given in the `deploy` directory.


## Prerequisites

You need the following:

- **Kubernetes 1.8+**
- **Sysdig Monitor** - see the [installation instructions][sysdig-monitor-docs-installation].
- **Sysdig Monitor API Token** - For more information, see [Sydig Monitor API Documentation][sysdig-monitor-docs-api]. Do not confuse the **API token** with the **agent access key**. They are not the same. API token is used by the Sysdig metrics server while accessing the API.

If you are using a Kubernetes whose version is between *>=1.8.0 and <1.11.0* you need to enable the `--horizontal-pod-autoscaler-use-rest-clients=true` flag in the `kube-controller-manager`. To check if you have this flag enabled in the `kube-controller-manager`, run this command:  

```
kubectl get pods `kubectl get pods -n kube-system | grep kube-controller-manager | grep Running | cut -d' ' -f 1` -n kube-system -o yaml
```

If the version you are using is >=1.11.0, you don't need to enable this flag, as it's enabled by default.
**Warning**: Setting the `--horizontal-pod-autoscaler-use-rest-clients=true` flag to `false` enables Heapster-based autoscaling, which is deprecated.

## Installation

The configuration files that you can find under the [deploy](./deploy) directory
are just for reference. Every deployment is unique so tweak them as needed. At
the very least, you need to use your own **access key** and **API token**.

In this example, you will deploy [kuard][kuard], a demo application found in the "Kubernetes Up and Running" book.

1. [02-sysdig-metrics-server.yml](./deploy/02-sysdig-metrics-server.yml) installs a `secret` in Kubernetes containing the Sysdig Monitor API token. Edit it to use your own token.


2. Deploy [kuard][kuard]. This application is deployed with three replicas by default.

   ```
   $ kubectl apply -f deploy/00-kuard.yml
   ```

3. Check if the replicas are running:

   ```
    $ kubectl get pods -l app=kuard -o wide
    NAME                    READY     STATUS    RESTARTS   AGE       IP          NODE
    kuard-bcc7bf7df-clv2f   1/1       Running   0          1m        10.46.0.2   worker-node-2
    kuard-bcc7bf7df-d9svn   1/1       Running   0          1m        10.40.0.2   worker-node-1
    kuard-bcc7bf7df-zg8nc   1/1       Running   0          1m        10.46.0.3   worker-node-2
    ```

4. Install Sysdig Monitor if you haven't done it yet. For more information, see [Installing Sysdig Monitor][sysdig-monitor-docs-installation].


5. Run the following to deploy a number of required objects like
   a custom namespace `custom-metrics`, required RBAC roles, permissions,
   bindings and the service object for the Sysdig metrics server:

   ```
   $ kubectl apply -f deploy/01-sysdig-metrics-rbac.yml
   ```

   It creates the following:

   The namespace that will contain the metrics server
   ```
   kind: Namespace
   apiVersion: v1
   metadata:
     name: custom-metrics
   ```

   The ServiceAccount for the metrics server. ServiceAccounts are a way to authenticate processes which run in pods.

   ```
   kind: ServiceAccount
   apiVersion: v1
   metadata:
     name: custom-metrics-apiserver
     namespace: custom-metrics
   ```

6. Create a ClusterRoleBinding that will bind the ClusterRole `system:auth-delegator` with the ServiceAccount `custom-metrics-api-server` you created to delegate authentication decisions to the Kubernetes API server.

   ```
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRoleBinding
   metadata:
     name: custom-metrics:system:auth-delegator
   roleRef:
     apiGroup: rbac.authorization.k8s.io
     kind: ClusterRole
     name: system:auth-delegator
   subjects:
   - kind: ServiceAccount
     name: custom-metrics-apiserver
     namespace: custom-metrics
   ```

7. Create a RoleBinding that will bind the Role that will authorize the application to access the `extension-apiserver-authentication` configmap.

   ```
   apiVersion: rbac.authorization.k8s.io/v1
   kind: RoleBinding
   metadata:
     name: custom-metrics-auth-reader
     namespace: kube-system
   roleRef:
     apiGroup: rbac.authorization.k8s.io
     kind: Role
     name: extension-apiserver-authentication-reader
   subjects:
   - kind: ServiceAccount
     name: custom-metrics-apiserver
     namespace: custom-metrics
   ```

8. Create a new ClusterRole that will have access to retrieve and list the namespaces, pods, and services.

   ```
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRole
   metadata:
     name: custom-metrics-resource-reader
   rules:
   - apiGroups:
     - ""
     resources:
     - namespaces
     - pods
     - services
     verbs:
     - get
     - list
   ```

9. Bind it with the service account you created for the metrics Server.

   ```
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRoleBinding
   metadata:
     name: custom-metrics-apiserver-resource-reader
   roleRef:
     apiGroup: rbac.authorization.k8s.io
     kind: ClusterRole
     name: custom-metrics-resource-reader
   subjects:
   - kind: ServiceAccount
     name: custom-metrics-apiserver
     namespace: custom-metrics
   ```

10. Create also a cluster role with complete access to the API group `custom.metrics.k8s.io` where you will publish the metrics.

   ```
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRole
   metadata:
     name: custom-metrics-getter
   rules:
   - apiGroups:
     - custom.metrics.k8s.io
     resources:
     - "*"
     verbs:
     - "*"
   ```

11. Bind it to the HPA so it can retrieve the metrics.

   ```
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRoleBinding
   metadata:
     name: hpa-custom-metrics-getter
   roleRef:
     apiGroup: rbac.authorization.k8s.io
     kind: ClusterRole
     name: custom-metrics-getter
   subjects:
   - kind: ServiceAccount
     name: horizontal-pod-autoscaler
     namespace: kube-system
   ```

12. Create the Service to publish the metrics:

   ```
   apiVersion: v1
   kind: Service
   metadata:
     name: api
     namespace: custom-metrics
   spec:
     ports:
     - port: 443
       targetPort: 443
     selector:
       app: custom-metrics-apiserver      
   ```

13. Create the API endpoint:

   ```
   apiVersion: apiregistration.k8s.io/v1beta1
   kind: APIService
   metadata:
     name: v1beta1.custom.metrics.k8s.io
   spec:
     insecureSkipTLSVerify: true
     group: custom.metrics.k8s.io
     groupPriorityMinimum: 1000
     versionPriority: 5
     service:
       name: api
       namespace: custom-metrics
     version: v1beta1
   ```

14. Deploy the metrics server, but first you need to create the secret with your API key:

   ```
   kubectl create secret generic --from-literal access-key=<YOUR_API_KEY> -n custom-metrics sysdig-api
   ```

15. Create the deployment with:

   ```
   $ kubectl apply -f deploy/02-sysdig-metrics-server.yml
   ```

   You can retrieve the full list of metrics available by using the following command:

   ```
   $ kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1" | jq -r ".resources[].name"
   ```

   If you want to know the value of a metric, run:

   ```
   $ kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/<NAMESPACE_NAME>/services/<SERVICE_NAME>/<METRIC_NAME>" | jq .
   ```

   For example:

   ```
   $ kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/default/services/kuard/net.http.request.count" | jq .
   {
     "kind": "MetricValueList",
     "apiVersion": "custom.metrics.k8s.io/v1beta1",
     "metadata": {
       "selfLink": "/apis/custom.metrics.k8s.io/v1beta1/namespaces/default/services/kuard/net.http.request.count"
     },
     "items": [
       {
         "describedObject": {
           "kind": "Service",
           "namespace": "default",
           "name": "kuard",
           "apiVersion": "/__internal"
         },
         "metricName": "net.http.request.count",
         "timestamp": "2019-02-20T21:53:22Z",
         "value": "0"
       }
     ]
   }
   ```

16. Deploy our custom autoscaler that scales the service based on the
   `net.http.request.count` metric.

   ```
   $ kubectl apply -f deploy/03-kuard-hpa.yml
   ```

At this point you should be able to see the autoscaler in action. In the
example, you set a threshold of 100 requests per minute.

17. To generate some traffic with [hey][hey]:

    $ hey -c 5 -q 85 -z 24h http://10.103.86.213

18. Use the following command to watch the autoscaler:

    $ kubectl get hpa kuard-autoscaler -w
    NAME               REFERENCE          TARGETS       MINPODS   MAXPODS   REPLICAS   AGE
    kuard-autoscaler   Deployment/kuard   105763m/100   3         10        8          2d

## Troubleshooting

If you encounter any problems that the documentation does not address,
[file an issue][new-issue].

## Contributing

Thanks for taking the time to join our community and start contributing!

- Please familiarize yourself with the [Code of Conduct][code-of-conduct]
  before contributing.
- See [CONTRIBUTING.md][contributing] for information about setting up your
  environment, the workflow that we expect, and instructions on the developer
  certificate of origin that we require.
- Check out the [issues][issues] and [our roadmap][roadmap].

## Relevant links

From the Kubernetes project:

- [Types in Custom Metrics API][l1]
- [Library for writing a Custom Metrics API server][l2]
- [Library for writing a Kubernetes-style API Server][l3]
- Documentation: [Horizontal Pod Autoscaling][l4]

Other links:

- [Sysdig's blog post about HPA][l5]
- [Kubernetes Prometheus Adapter][l6]
- [kubeadm workshop][l7]
- [Basic Custom Metrics API server using minikube][l8]

## Credits and license

This project wouldn't be possible without the great job done by the community. See the
links above for some of the things I've been using in different ways.

The tagging policy and the contributing guide in this project is based on
Contour's.

This project is open source and it uses a [permissive license][license].

[l1]: https://github.com/kubernetes/metrics/tree/master/pkg/apis/custom_metrics
[l2]: https://github.com/kubernetes-incubator/custom-metrics-apiserver
[l3]: https://github.com/kubernetes/apiserver
[l4]: https://github.com/kubernetes/community/blob/master/contributors/design-proposals/autoscaling/horizontal-pod-autoscaler.md
[l5]: https://sysdig.com/blog/kubernetes-scaler/
[l6]: https://github.com/directXMan12/k8s-prometheus-adapter/
[l7]: https://github.com/luxas/kubeadm-workshop
[l8]: https://github.com/vishen/k8s-custom-metrics

[1]: https://travis-ci.org/sevein/k8s-sysdig-adapter.svg?branch=master
[2]: https://travis-ci.org/sevein/k8s-sysdig-adapter

[new-issue]: https://github.com/draios/kubernetes-sysdig-metrics-apiserver/issues/new
[roadmap]: https://github.com/draios/kubernetes-sysdig-metrics-apiserver/milestones
[issues]: https://github.com/draios/kubernetes-sysdig-metrics-apiserver/issues
[contributing]: /CONTRIBUTING.md
[code-of-conduct]: /CODE_OF_CONDUCT.md
[license]: /LICENSE
[kuard]: https://github.com/kubernetes-up-and-running/kuard
[custom-metrics-api-types]: https://github.com/kubernetes/metrics/tree/master/pkg/apis/custom_metrics
[sysdig-monitor]: https://sysdig.com/product/monitor/
[sysdig-monitor-docs-installation]: https://sysdigdocs.atlassian.net/wiki/spaces/Platform/pages/190775522/Agent+Installation
[sysdig-monitor-docs-api]: https://support.sysdig.com/hc/en-us/articles/205233166
[minikube]: https://github.com/kubernetes/minikube
[hey]: https://github.com/rakyll/hey
