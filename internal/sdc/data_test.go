package sdc

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

var (
	metricsJSONResponse = `
{
  "net.http.request.count": {
    "aggregationForGroup": "sum",
    "aggregations": [
      "timeAvg",
      "max",
      "sum",
      "min",
      "avg"
    ],
    "avg": true,
    "canCalculate": true,
    "canFilter": false,
    "canGroupBy": false,
    "canMonitor": false,
    "category": "network",
    "concat": false,
    "context": [],
    "count": false,
    "countDistinct": false,
    "description": "Count ofr HTTP requests",
    "groupAggregations": [
      "max",
      "sum",
      "min",
      "avg"
    ],
    "groupBy": [
      "kubernetes.replicaSet.label.k8s-app",
      "kubernetes.deployment.name",
      "kubernetes.node.label.kubernetes.io/hostname",
      "kubernetes.replicaSet.label",
      "container.label.org.label-schema.name",
      "kubernetes.service.label",
      "host.mac",
      "container.id",
      "container.name",
      "kubernetes.replicaSet.label.pod-template-hash",
      "kubernetes.daemonSet.label.app",
      "kubernetes.pod.label.controller-revision-hash",
      "net.http.url",
      "container.label.io.kubernetes.container.name",
      "cloudProvider.availabilityZone",
      "container.image.id",
      "kubernetes.namespace.name",
      "kubernetes.daemonSet.label",
      "kubernetes.service.name",
      "ecs.serviceName",
      "kubernetes.pod.label.app",
      "kubernetes.service.label.app",
      "kubernetes.pod.label.k8s-app",
      "container.label.io.kubernetes.pod.name",
      "agent.tag",
      "kubernetes.deployment.label.k8s-app",
      "kubernetes.pod.label.pod-template-hash",
      "container.label",
      "kubernetes.node.name",
      "kubernetes.pod.label",
      "ecs.clusterName",
      "ecs.taskFamilyName",
      "container.label.org.label-schema.url",
      "kubernetes.replicationController.name",
      "kubernetes.pod.label.name",
      "net.http.statusCode",
      "kubernetes.replicaSet.label.app",
      "kubernetes.daemonSet.label.k8s-app",
      "container.label.org.label-schema.vendor",
      "kubernetes.pod.label.pod-template-generation",
      "kubernetes.pod.name",
      "cloudProvider.tag",
      "cloudProvider.region",
      "net.http.method",
      "kubernetes.replicaSet.name",
      "container.label.io.kubernetes.pod.namespace",
      "container.label.org.label-schema.vcs-url",
      "kubernetes.daemonSet.label.name",
      "kubernetes.service.label.kubernetes.io/cluster-service",
      "cloudProvider.resource.name",
      "kubernetes.deployment.label.app",
      "kubernetes.service.label.k8s-app",
      "kubernetes.node.label.beta.kubernetes.io/arch",
      "kubernetes.service.label.kubernetes.io/name",
      "container.label.maintainer",
      "kubernetes.deployment.label",
      "kubernetes.node.label.beta.kubernetes.io/os",
      "container.label.works.weave.role",
      "kubernetes.daemonSet.name",
      "host.hostName",
      "kubernetes.node.label",
      "container.image"
    ],
    "hidden": false,
    "id": "net.http.request.count",
    "identity": false,
    "max": true,
    "metricType": "counter",
    "min": true,
    "name": "HTTP request count",
    "namespaces": [
      "host",
      "host.container",
      "host.process",
      "cloudProvider",
      "kubernetes.cluster",
      "kubernetes.namespace",
      "kubernetes.deployment",
      "kubernetes.job",
      "kubernetes.daemonSet",
      "kubernetes.service",
      "kubernetes.node",
      "kubernetes.replicaSet",
      "kubernetes.pod",
      "mesos",
      "swarm",
      "ecs",
      "host.net"
    ],
    "quantity": "",
    "scopes": [
      "host"
    ],
    "shortName": "http count",
    "shorterName": "",
    "sum": true,
    "timeAggregations": [
      "timeAvg",
      "max",
      "sum",
      "min",
      "avg"
    ],
    "timeAvg": true,
    "type": "int",
    "unitPostfix": ""
  }
}`

	getJSONResponse = `
{
  "data": [
    {
      "d": [0.127],
      "t": 1523864340
    },
    {
      "d": [0.128],
      "t": 1523864350
    }
  ],
  "start": 1523864330,
  "end": 1523864350
}`
)

func TestData_Metrics(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/data/metrics", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, metricsJSONResponse)
	})

	payload, _, err := client.Data.Metrics(ctx)
	if err != nil {
		t.Errorf("Data.Metrics returned error: %v", err)
	}

	if have, want := len(payload), 1; have != want {
		t.Errorf("Data.Metrics returned %d items, expected %d", have, want)
	}
	metricInfo := payload["net.http.request.count"]
	if have, want := metricInfo.Type, "int"; have != want {
		t.Errorf("Data.Metrics returned type %s, expected %s", have, want)
	}
}

func TestData_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/data/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, getJSONResponse)
	})

	req := &GetDataRequest{Last: -20, Sampling: 10}
	req = req.WithMetric("memory.used.percent", &MetricAggregation{Time: "timeAvg", Group: "avg"})
	payload, _, err := client.Data.Get(ctx, req)
	if err != nil {
		t.Errorf("Data.Get returned error: %v", err)
	}

	if have, want := len(payload.Samples), 2; have != want {
		t.Errorf("Data.Get returned %d samples, expected %d", have, want)
	}
	if have, want := payload.Start, Timestamp(time.Unix(1523864330, 0)); have != want {
		t.Fatalf("Data.Get returned %s, expected %s", have.String(), want.String())
	}
	if have, want := payload.End, Timestamp(time.Unix(1523864350, 0)); have != want {
		t.Fatalf("Data.Get returned %s, expected %s", have.String(), want.String())
	}
}
