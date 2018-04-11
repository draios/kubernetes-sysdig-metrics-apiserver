package cmprovider

import (
	"time"

	"github.com/golang/glog"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/metrics/pkg/apis/custom_metrics"

	// Temporar hack until I can vendor it.
	_cma_provider "github.com/sevein/k8s-sysdig-adapter/internal/custom-metrics-apiserver/pkg/provider"
)

type sysdigProvider struct {
	mapper       apimeta.RESTMapper
	kubeClient   dynamic.ClientPool
	rateInterval time.Duration
}

func NewSysdigProvider(mapper apimeta.RESTMapper, kubeClient dynamic.ClientPool, labelPrefix string, updateInterval time.Duration, rateInterval time.Duration, stopChan <-chan struct{}) _cma_provider.CustomMetricsProvider {
	lister := &cachingMetricsLister{
		updateInterval: updateInterval,
	}
	lister.RunUntil(stopChan)
	return &sysdigProvider{
		mapper:       mapper,
		kubeClient:   kubeClient,
		rateInterval: rateInterval,
	}
}

func (p *sysdigProvider) getSingle(info _cma_provider.CustomMetricInfo, namespace, name string) (*custom_metrics.MetricValue, error) {
	return &custom_metrics.MetricValue{}, nil
}

func (p *sysdigProvider) getMultiple(info _cma_provider.CustomMetricInfo, namespace, name string) (*custom_metrics.MetricValueList, error) {
	return &custom_metrics.MetricValueList{}, nil
}

func (p *sysdigProvider) GetRootScopedMetricByName(groupResource schema.GroupResource, name string, metricName string) (*custom_metrics.MetricValue, error) {
	glog.V(10).Infof("GetRootScopedMetricByName...")
	info := _cma_provider.CustomMetricInfo{
		GroupResource: groupResource,
		Metric:        metricName,
		Namespaced:    false,
	}
	return p.getSingle(info, "", name)
}

func (p *sysdigProvider) GetRootScopedMetricBySelector(groupResource schema.GroupResource, selector labels.Selector, metricName string) (*custom_metrics.MetricValueList, error) {
	glog.V(10).Infof("GetRootScopedMetricBySelector...")
	info := _cma_provider.CustomMetricInfo{
		GroupResource: groupResource,
		Metric:        metricName,
		Namespaced:    false,
	}
	return p.getMultiple(info, "", selector.String())
}

func (p *sysdigProvider) GetNamespacedMetricByName(groupResource schema.GroupResource, namespace string, name string, metricName string) (*custom_metrics.MetricValue, error) {
	glog.V(10).Infof("GetNamespacedMetricByName...")
	info := _cma_provider.CustomMetricInfo{
		GroupResource: groupResource,
		Metric:        metricName,
		Namespaced:    true,
	}
	return p.getSingle(info, namespace, name)
}

func (p *sysdigProvider) GetNamespacedMetricBySelector(groupResource schema.GroupResource, namespace string, selector labels.Selector, metricName string) (*custom_metrics.MetricValueList, error) {
	glog.V(10).Infof("GetNamespacedMetricBySelector...")
	info := _cma_provider.CustomMetricInfo{
		GroupResource: groupResource,
		Metric:        metricName,
		Namespaced:    true,
	}
	return p.getMultiple(info, namespace, selector.String())
}

func (p *sysdigProvider) ListAllMetrics() []_cma_provider.CustomMetricInfo {
	glog.V(10).Infof("ListAllMetrics...")
	return []_cma_provider.CustomMetricInfo{}
}

type cachingMetricsLister struct {
	updateInterval time.Duration
}

func (l *cachingMetricsLister) Run() {
	l.RunUntil(wait.NeverStop)
}

func (l *cachingMetricsLister) RunUntil(stopChan <-chan struct{}) {
	go wait.Until(func() {
		if err := l.updateMetrics(); err != nil {
			utilruntime.HandleError(err)
		}
	}, l.updateInterval, stopChan)
}

func (l *cachingMetricsLister) updateMetrics() error {
	glog.V(10).Infof("Set available metric list from Sysdig...")
	return nil
}
