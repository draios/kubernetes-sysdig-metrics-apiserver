package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/util/logs"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/sevein/k8s-sysdig-adapter/internal/cmprovider"

	// Temporar hack until I can vendor it.
	_cma_server "github.com/sevein/k8s-sysdig-adapter/internal/custom-metrics-apiserver/pkg/cmd/server"
	_cma_dynamicmapper "github.com/sevein/k8s-sysdig-adapter/internal/custom-metrics-apiserver/pkg/dynamicmapper"
)

// This is the name associated to our CustomMetricsAdapterServer.
const customMetricAdapterName = "sysdig-custom-metrics-adapter"

func main() {
	// Workaround for this issue: https://github.com/kubernetes/kubernetes/issues/17162.
	flag.CommandLine.Parse([]string{})

	logs.InitLogs()
	defer logs.FlushLogs()

	cmd := command(os.Stdout, os.Stderr, wait.NeverStop)
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func command(out, errOut io.Writer, stopCh <-chan struct{}) *cobra.Command {
	baseOpts := _cma_server.NewCustomMetricsAdapterServerOptions(out, errOut)
	o := adapterOpts{
		CustomMetricsAdapterServerOptions: baseOpts,
	}

	cmd := &cobra.Command{
		Short: "Launch the custom metrics API adapter server",
		Long:  "Launch the custom metrics API adapter server",
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			return o.runCustomMetricsAdapterServer(stopCh)
		},
	}

	flags := cmd.Flags()
	o.SecureServing.AddFlags(flags)
	o.Authentication.AddFlags(flags)
	o.Authorization.AddFlags(flags)
	o.Features.AddFlags(flags)

	flags.StringVar(&o.RemoteKubeConfigFile, "lister-kubeconfig", o.RemoteKubeConfigFile,
		"kubeconfig file pointing at the 'core' kubernetes server with enough rights to list any described objets")
	flags.DurationVar(&o.MetricsRelistInterval, "metrics-relist-interval", o.MetricsRelistInterval,
		"interval at which to re-list the set of all available metrics from Sysdig")
	flags.DurationVar(&o.RateInterval, "rate-interval", o.RateInterval,
		"period of time used to calculate rate metrics from cumulative metrics")
	flags.DurationVar(&o.DiscoveryInterval, "discovery-interval", o.DiscoveryInterval,
		"interval at which to refresh API discovery information")
	flags.StringVar(&o.LabelPrefix, "label-prefix", o.LabelPrefix,
		"prefix to expect on labels referring to pod resources. For example, if the prefix is 'kube_', any series with the 'kube_pod' label would be considered a pod metric")

	return cmd
}

type adapterOpts struct {
	*_cma_server.CustomMetricsAdapterServerOptions

	// RemoteKubeConfigFile is the config used to list pods from the master API server
	RemoteKubeConfigFile string

	// MetricsRelistInterval is the interval at which to relist the set of available metrics
	MetricsRelistInterval time.Duration

	// RateInterval is the period of time used to calculate rate metrics
	RateInterval time.Duration

	// DiscoveryInterval is the interval at which discovery information is refreshed
	DiscoveryInterval time.Duration

	// LabelPrefix is the prefix to expect on labels for Kubernetes resources
	// (e.g. if the prefix is "kube_", we'd expect a "kube_pod" label for pod metrics).
	LabelPrefix string
}

// runCustomMetricsAdapterServer runs our CustomMetricsAdapterServer.
func (o adapterOpts) runCustomMetricsAdapterServer(stopCh <-chan struct{}) error {
	config, err := o.Config()
	if err != nil {
		fmt.Println(err)
		return err
	}

	config.GenericConfig.EnableMetrics = true

	var clientConfig *rest.Config
	if len(o.RemoteKubeConfigFile) > 0 {
		loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: o.RemoteKubeConfigFile}
		loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
		clientConfig, err = loader.ClientConfig()
	} else {
		clientConfig, err = rest.InClusterConfig()
	}
	if err != nil {
		return fmt.Errorf("unable to construct lister client config to initialize provider: %v", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(clientConfig)
	if err != nil {
		return fmt.Errorf("unable to construct discovery client for dynamic client: %v", err)
	}

	dynamicMapper, err := _cma_dynamicmapper.NewRESTMapper(discoveryClient, apimeta.InterfacesForUnstructured, o.DiscoveryInterval)
	if err != nil {
		return fmt.Errorf("unable to construct dynamic discovery mapper: %v", err)
	}

	clientPool := dynamic.NewClientPool(clientConfig, dynamicMapper, dynamic.LegacyAPIPathResolverFunc)
	if err != nil {
		return fmt.Errorf("unable to construct lister client to initialize provider: %v", err)
	}

	server, err := config.Complete().New(
		// Name of the CustomMetricsAdapterServer (for logging purposes).
		customMetricAdapterName,
		// CustomMetricsProvider.
		cmprovider.NewSysdigProvider(dynamicMapper, clientPool, o.LabelPrefix, o.MetricsRelistInterval, o.RateInterval, stopCh),
		// ExternalMetricsProvider (which we're not implementing)
		nil,
	)
	if err != nil {
		return err
	}

	return server.GenericAPIServer.PrepareRun().Run(stopCh)
}
