package main

import (
	metrics "code.cloudfoundry.org/go-metric-registry"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	loggr := log.New(os.Stderr, "[metrics registry] ", 0)
	registry := metrics.NewRegistry(
		loggr,

		// If not provided, the registry will try to listen on an existing http server
		// Typically metrics.WithTLSServer should be used
		metrics.WithServer(0),
	)

	cntr := registry.NewCounter("counter_name", "This is an example counter. Use this area to give a useful description about your counter.")
	gauge := registry.NewGauge("gauge_name", "This is an example gauge. Use this area to give a useful description about your gauge.")

	cntr.Add(10)
	gauge.Set(38)

	// Calling NewMetric with the same name, help text, and tags will return the existing metric
	cntr2 := registry.NewCounter("counter_name", "This is an example counter. Use this area to give a useful description about your counter.")
	cntr2.Add(15) // Endpoint will show 25 as the value of counter_name

	waitForTermination()
}

func waitForTermination() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	<-c
}