# Kubernetes Custom Metrics Adapter for Sysdig

[![Build status][1]][2]

Table of contents:

- [Introduction](#introduction)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Playground](#playground)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [Relevant links](#relevant-links)
- [Credits and License](#credits-and-license)

## Introduction

`k8s-sysdig-adapter` is an implementation of the
[Custom Metrics API][custom-metrics-api-types] using
[Sysdig Monitor][sysdig-monitor].

If you have a Kubernetes cluster and you are a Sysdig user, this adapter
enables you to create [horizontal pod autoscalers][l4] based on metrics provided
by Sysdig's monitoring solution.

In the following example, we're creating an autoscaler based on the
`net.http.request.count` metric. The autoscaler will adjust the number of pods
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
        name: kuard
      metricName: net.http.request.count
      targetValue: 100
```

## Prerequisites

You're going to need:

- **Kubernetes 1.8+**
- **Sysdig Monitor** - see the [installation instructions][sysdig-monitor-docs-installation].
- **Sysdig Monitor Access Key** - which you've used during the installation of Sysdig Monitor.
- **Sysdig Monitor API Token** - see where to find it in [these instructions][sysdig-monitor-docs-api]. Do not confuse the **API token** with the **agent access key**, they're not the same! This is the API token that our metrics server is going to use when accessing the API.

## Installation

The configuration files that you can find under the [deploy](./deploy) directory
are just for reference. Every deployment is unique so tweak them as needed. At
the very least, you need to use your own **access key** and **API token** as
follows:

- [01-sysdig-daemon-set.yml](./deploy/01-sysdig-daemon-set.yml) installs the
  Sysdig agent - edit the `secret` so it uses your own access key.
- [03-sysdig-metrics-server.yml](./deploy/03-sysdig-metrics-server.yml) installs
  a `secret` in Kubernetes containing the Sysdig Monitor API token - edit it to
  use your own token.

Now we're ready to start! :tada:

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

2. Install Sysdig Monitor if you haven't done it yet - they have
   [great docs][sysdig-monitor-docs-installation] that you can use. In this
   example, we're going to install it using a `DaemonSet` object as follows:

   ```
   $ kubectl apply -f deploy/01-sysdig-daemon-set.yml
   ```

   **Don't forget to add your own Agent key (not the API token) to the file!**

3. The following command is going to deploy a number of required objects like
   a custom namespace `custom-metrics`, required RBAC roles, permissions,
   bindings and the service object for our metrics server:

   ```
   $ kubectl apply -f deploy/02-sysdig-metrics-rbac.yml
   ```

4. Deploy the metrics server with:

   ```
   $ kubectl apply -f deploy/03-sysdig-metrics-server.yml
   ```

   **Don't forget to add your own API token to the file!**

   It should be possible to retrieve the full list of metrics available using
   the following command:

   ```
   $ kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1" | jq -r ".resources[].name"
   ```

5. Deploy our custom autoscaler that scales our service based on the
   `net.http.request.count` metric.

   ```
   $ kubectl apply -f deploy/04-kuard-hpa.yml
   ```

At this point you should be able to see the autoscaler in action. In the
example, we set a threshold of 100 requests per minute. Let's generate some
traffic with [hey][hey]:

    $ hey -c 5 -q 85 -z 24h http://10.103.86.213

Finally, use the following command to watch the autoscaler:

    $ kubectl get hpa kuard-autoscaler -w
    NAME               REFERENCE          TARGETS       MINPODS   MAXPODS   REPLICAS   AGE
    kuard-autoscaler   Deployment/kuard   105763m/100   3         10        8          2d

## Playground

You can find a playground based on Vagrant virtual machines under the
[playground](./playground) directory in this repository. You can use it to demo
this project or for development purposes.

[minikube][minikube] has not been tried yet! See
[issue #3](https://github.com/draios/kubernetes-sysdig-metrics-apiserver/issues/3) for more
details.

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

This project wouldn't be possible without the great job done by others. See the
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
[sysdig-monitor-docs-installation]: https://support.sysdig.com/hc/en-us/articles/206770633-Sysdig-Install-Kubernetes-
[sysdig-monitor-docs-api]: https://support.sysdig.com/hc/en-us/articles/205233166
[minikube]: https://github.com/kubernetes/minikube
[hey]: https://github.com/rakyll/hey
