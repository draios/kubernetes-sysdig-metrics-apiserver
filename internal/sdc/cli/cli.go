package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/sevein/k8s-sysdig-adapter/internal/sdc"
)

var (
	client  *sdc.Client
	rootCmd = &cobra.Command{Use: "sdc-cli"}
	ctx     = context.Background()
)

func main() {
	if err := run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func run() error {
	var token = os.Getenv("SDC_TOKEN")
	if token == "" {
		return errors.New("token not provided, use environment SDC_TOKEN")
	}
	client = sdc.NewClient(nil, token)

	rootCmd.AddCommand(newGetDataCmd(os.Stdout))
	rootCmd.AddCommand(newListMetricsCmd(os.Stdout))
	return rootCmd.Execute()
}

func newGetDataCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use: "get-data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetData(out, cmd)
		},
	}
	cmd.Flags().String("metric", "", "Name of the metric")
	return cmd
}

func runGetData(out io.Writer, cmd *cobra.Command) error {
	metric, _ := cmd.Flags().GetString("metric")
	if metric == "" {
		return errors.New("metric name is empty")
	}
	req := &sdc.GetDataRequest{
		DataSourceType: "host",
		Last:           600,
		Sampling:       60,
		Metrics: []sdc.Metric{
			sdc.Metric{
				ID:           metric,
				Aggregations: sdc.MetricAggregation{Time: "timeAvg", Group: "avg"},
			},
		},
	}
	data, _, err := client.Data.Get(ctx, req)
	if err != nil {
		return err
	}
	for _, item := range data.Data {
		fmt.Fprintf(out, "Data point: %f (%s)\n", item.Points, item.Time.String())
	}
	return nil
}

func newListMetricsCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use: "list-metrics",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListMetrics(out, cmd)
		},
	}
	return cmd
}

func runListMetrics(out io.Writer, cmd *cobra.Command) error {
	metrics, _, err := client.Data.Metrics(ctx)
	if err != nil {
		return err
	}
	for id, metric := range *metrics {
		fmt.Fprintf(out, "Metric name: %s, type: %s\n", id, metric.Type)
	}
	return nil
}
