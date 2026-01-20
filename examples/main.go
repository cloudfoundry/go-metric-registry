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
	cntrVec := registry.NewCounterVec("counter_vector_name", "This is an example counter vector. Use this area to give a useful description about your counter vector.", []string{"status", "app"})
	gauge := registry.NewGauge("gauge_name", "This is an example gauge. Use this area to give a useful description about your gauge.")
	gaugeVec := registry.NewGaugeVec("gauge_vector_name", "This is an example gauge vector. Use this area to give a useful description about your gauge vector.", []string{"state", "source"})
	hist := registry.NewHistogram("histogram_name", "This is an example histogram. Use this area to give a useful description about your histogram.", []float64{1.0})
	histVec := registry.NewHistogramVec("histogram_vector_name", "This is an example histogram vector. Use this area to give a useful description about your histogram vector.", []string{"mode"}, []float64{10, 50, 100})

	cntr.Add(10)
	cntrVec.Add(0.25, []string{"success", "IntranetPortal"}) // Define label names when you initialize the vector and label values here, when you add new value to it
	gauge.Set(38)
	gaugeVec.Set(12.5, []string{"active", "MobileApp"})
	gaugeVec.Add(0.5, []string{"active", "MobileApp"}) // Define label names when you initialize the vector and label values here, when you add new value to it
	hist.Observe(98.01234)
	histVec.Observe(67, []string{"detached"}) // Define label names when you initialize the vector and label values here, when you add new value to it

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
