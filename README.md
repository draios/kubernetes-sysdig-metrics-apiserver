# Kubernetes Custom Metrics Adapter for Sysdig

[![Build status][1]][2]

Table of contents:

- [Introduction](#introduction)
- [Installation](#installation)
- [Playground](#playground)
- [Relevant links](#relevant-links)

## Introduction

`k8s-sysdig-adapter` is an implementation of the
[Custom Metrics API][custom-metrics-api-types] using
[Sysdig Monitor][sysdig-monitor].

Essentially, this component is a custom Kubernetes API server that queries
Sysdig Monitor's API for metrics data and exposes it to Kubernetes. You can
think of it as a channel adapter between Sysdig and the
[Horizontal Pod Autoscaling API][hpa] for Kubernetes.

Once it's installed you should be able to deploy `HorizontalPodAutoscaler`
objects (also known as autoscalers) fed with metrics provided by Sysdig Monitor.
The autoscaler object must use the `autoscaling/v2beta1` form like in the
following example:

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
        name: kuard
      metricName: net.http.request.count
      targetValue: 100
```

This autoscaler is based on the `net.http.request.count` metric. The autoscaler
will adjust the number of pods deployed as the metric fluctuates over or below
the threshold (in the example, 100 reqs/min).

## Download

`k8s-sysdig-adapter` is distributed as a Docker image.

You can find it at `sevein/k8s-sysdig-adapter:latest`.

## Installation

Use these instructions only as a reference. Every deployment is unique!

1. For the purpose of this example we're going to deploy [kuard][kuard], a demo
   application found in the "Kubernetes Up and Running" book. This application
   is deployed with three replicas by default.

   ```
   $ kubectl apply -f deploy/00-kuard.yml
   ```

   Let's check that it's running:

   ```
    $ kubectl get pods -l app=kuard -o wide
    NAME                    READY     STATUS    RESTARTS   AGE       IP          NODE
    kuard-bcc7bf7df-clv2f   1/1       Running   0          1m        10.46.0.2   worker-node-2
    kuard-bcc7bf7df-d9svn   1/1       Running   0          1m        10.40.0.2   worker-node-1
    kuard-bcc7bf7df-zg8nc   1/1       Running   0          1m        10.46.0.3   worker-node-2
    ```

2. [Install the Sysdig Monitor agent][sysdig-monitor-inst-docs]. It's deployed
   as a DaemonSet. Make sure that you include in the documet your own agent
   access key.

   ```
   $ kubectl apply -f deploy/01-sysdig-daemon-set.yml
   ```

3. The following command is going to deploy the required RBAC roles,
   permissions and bindings. It uses the namespace `custom-metrics`.

   ```
   $ kubectl apply -f deploy/02-sysdig-metrics-rbac.yml
   ```

4. Record the base64-encoded version of your agent key.<br />
   Let's assume that our key is `59493980-bbab-44e5-81b2-d80d59192fcd`.

   ```
   $ echo -n 59493980-bbab-44e5-81b2-d80d59192fcd | base64
   NTk0OTM5ODAtYmJhYi00NGU1LTgxYjItZDgwZDU5MTkyZmNk
   ```

5. Deploy the metrics server. Make sure that the file is edited as needed, e.g.
   target your own service and deployment and use your own base64-encoded agent
   access key that we've generated in the previous step.

   ```
   $ kubectl apply -f deploy/03-sysdig-metrics-server.yml
   ```

6. Finally, deploy the autoscaler targeting our `kuard` service.

   ```
   $ kubectl apply -f deploy/04-kuard-hpa.yml
   ```

## Playground

You can find a playground based on Vagrant virtual machines under the
[playground](./playground) directory in this repository. You can use it to demo
this project or for development purposes.

[minikube][minikube] has not been tried yet! See
[issue #3](https://github.com/sevein/k8s-sysdig-adapter/issues/3) for more
details.

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

[l1]: https://github.com/kubernetes/metrics/tree/master/pkg/apis/custom_metrics
[l2]: https://github.com/kubernetes-incubator/custom-metrics-apiserver
[l3]: https://github.com/kubernetes/apiserver
[l4]: https://github.com/kubernetes/community/blob/master/contributors/design-proposals/autoscaling/horizontal-pod-autoscaler.md
[l5]: https://sysdig.com/blog/kubernetes-scaler/
[l6]: https://github.com/directXMan12/k8s-prometheus-adapter/
[l7]: https://github.com/luxas/kubeadm-workshop

[1]: https://travis-ci.org/sevein/k8s-sysdig-adapter.svg?branch=master
[2]: https://travis-ci.org/sevein/k8s-sysdig-adapter

[kuard]: https://github.com/kubernetes-up-and-running/kuard
[custom-metrics-api-types]: https://github.com/kubernetes/metrics/tree/master/pkg/apis/custom_metrics
[hpa]: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.10/#horizontalpodautoscaler-autoscaling-v2beta1-
[sysdig-monitor]: https://sysdig.com/product/monitor/
[sysdig-monitor-inst-docs]: https://support.sysdig.com/hc/en-us/articles/206770633-Sysdig-Install-Kubernetes-
[minikube]: https://github.com/kubernetes/minikube
