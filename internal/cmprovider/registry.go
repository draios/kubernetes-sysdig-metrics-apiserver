package cmprovider

import (
	"sync"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/runtime/schema"

	// TODO: Vendor this
	cmaprovider "github.com/draios/kubernetes-sysdig-metrics-apiserver/internal/custom-metrics-apiserver/pkg/provider"

	"github.com/draios/kubernetes-sysdig-metrics-apiserver/internal/sdc"
)

type MetricsRegistry interface {
	UpdateMetrics(sdc.Metrics)
	Metric(name string) (metric *sdc.MetricDefinition, found bool)
	ListAllMetrics() []cmaprovider.CustomMetricInfo
}

type registry struct {
	mu sync.RWMutex

	// Map of metrics indexed by its names, e.g. net.http.request.count.
	defs map[string]*sdc.MetricDefinition

	// List metrics that we return to Kubernetes.
	metrics []cmaprovider.CustomMetricInfo
}

var _ MetricsRegistry = &registry{}

var acceptedSysdigMetricNamespaces = []string{
	"kubernetes.cluster",
	"kubernetes.namespace",
	"kubernetes.deployment",
	"kubernetes.job",
	"kubernetes.daemonSet",
	"kubernetes.service",
	"kubernetes.node",
	"kubernetes.replicaSet",
	"kubernetes.pod",
}

func wantedNamespace(n string) bool {
	for _, wanted := range acceptedSysdigMetricNamespaces {
		if wanted == n {
			return true
		}
	}
	return false
}

func hasNamespace(namespaces []string, wanted string) bool {
	for _, item := range namespaces {
		if item == wanted {
			return true
		}
	}
	return false
}

func (r *registry) UpdateMetrics(m sdc.Metrics) {
	newDefs := make(map[string]*sdc.MetricDefinition)
	for name, metric := range m {
		// Ignore non-quantifiable metrics.
		if metric.MetricType != "gauge" && metric.MetricType != "counter" {
			continue
		}
		// Only services for now.
		if hasNamespace(metric.Namespaces, "kubernetes.service") {
			newDefs[name] = metric
		}
	}
	newMetrics := make([]cmaprovider.CustomMetricInfo, 0, len(newDefs))
	for name := range newDefs {
		newMetrics = append(newMetrics, cmaprovider.CustomMetricInfo{
			GroupResource: schema.GroupResource{Resource: "services"},
			Metric:        name,
			Namespaced:    true,
		})
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defs = newDefs
	r.metrics = newMetrics
}

func (r *registry) Metric(name string) (*sdc.MetricDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	metric, ok := r.defs[name]
	if !ok {
		glog.V(10).Infof("metric %s not registered", name)
		return nil, false
	}
	return metric, true
}

func (r *registry) ListAllMetrics() []cmaprovider.CustomMetricInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.metrics
}
