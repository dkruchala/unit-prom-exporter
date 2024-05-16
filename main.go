package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dkruchala/unit-prom-exporter/unit"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var serverPort string = "9090"
var controlApi = unit.NewControlApiConnection()

type ApplicationsDescriptors map[string]applicationDescriptors

type unitCollector struct {
	ConnectionsAcceptedTotal *prometheus.Desc
	ConnectionsActive        *prometheus.Desc
	ConnectionsIdle          *prometheus.Desc
	ConnectionsClosedTotal   *prometheus.Desc
	RequestsTotal            *prometheus.Desc
	Applications             ApplicationsDescriptors
}

func (collector *unitCollector) register() {
	prometheus.MustRegister(collector)
}

type applicationDescriptors struct {
	ProcessRunning  *prometheus.Desc
	ProcessStarting *prometheus.Desc
	ProcessIdle     *prometheus.Desc
	RequestsActive  *prometheus.Desc
}

func newUnitCollector(metrics unit.UnitMetrics) *unitCollector {
	var collector unitCollector
	collector.ConnectionsAcceptedTotal = prometheus.NewDesc("unit_connections_accepted_total",
		"Shows total count of accepted connections",
		nil, nil,
	)
	collector.ConnectionsActive = prometheus.NewDesc("unit_connections_active",
		"Shows current count of active connections",
		nil, nil,
	)
	collector.ConnectionsIdle = prometheus.NewDesc("unit_connections_idle",
		"Shows current count of idle connections",
		nil, nil,
	)
	collector.ConnectionsClosedTotal = prometheus.NewDesc("unit_connections_closed_total",
		"Shows total count of closed connections",
		nil, nil,
	)
	collector.RequestsTotal = prometheus.NewDesc("unit_requests_total",
		"Shows total count of requests",
		nil, nil,
	)

	collector.Applications = make(ApplicationsDescriptors)
	for k := range metrics.Applications {
		collector.Applications[k] = applicationDescriptors{
			ProcessRunning: prometheus.NewDesc("unit_"+k+"_process_running",
				"Shows current count of running processes",
				nil, nil,
			),
			ProcessStarting: prometheus.NewDesc("unit_"+k+"_process_starting",
				"Shows current count of starting processes",
				nil, nil,
			),
			ProcessIdle: prometheus.NewDesc("unit_"+k+"_process_idle",
				"Shows current count of idle processes",
				nil, nil,
			),
			RequestsActive: prometheus.NewDesc("unit_"+k+"_requests_active",
				"Shows current count of requests",
				nil, nil,
			),
		}
	}

	return &collector
}

func (collector *unitCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.ConnectionsAcceptedTotal
	ch <- collector.ConnectionsActive
	ch <- collector.ConnectionsIdle
	ch <- collector.ConnectionsClosedTotal
	ch <- collector.RequestsTotal

	for _, v := range collector.Applications {
		ch <- v.ProcessIdle
		ch <- v.ProcessRunning
		ch <- v.ProcessStarting
		ch <- v.ProcessStarting
	}
}

func (collector *unitCollector) Collect(ch chan<- prometheus.Metric) {

	metrics, err := controlApi.GetStatus()
	if err != nil {
		fmt.Println("Cannot get metrics from Unit API")
		return
	}

	connectionsAcceptedTotal := prometheus.MustNewConstMetric(
		collector.ConnectionsAcceptedTotal,
		prometheus.CounterValue,
		metrics.Connections.Accepted,
	)
	connectionsActive := prometheus.MustNewConstMetric(
		collector.ConnectionsActive,
		prometheus.GaugeValue,
		metrics.Connections.Active,
	)
	connectionsIdle := prometheus.MustNewConstMetric(
		collector.ConnectionsIdle,
		prometheus.GaugeValue,
		metrics.Connections.Idle,
	)
	connectionsClosedTotal := prometheus.MustNewConstMetric(
		collector.ConnectionsClosedTotal,
		prometheus.CounterValue,
		metrics.Connections.Closed,
	)
	requestsTotal := prometheus.MustNewConstMetric(
		collector.RequestsTotal,
		prometheus.CounterValue,
		metrics.Requests.Total,
	)
	ch <- connectionsAcceptedTotal
	ch <- connectionsActive
	ch <- connectionsIdle
	ch <- connectionsClosedTotal
	ch <- requestsTotal

	for k, v := range collector.Applications {

		processRunning := prometheus.MustNewConstMetric(
			v.ProcessRunning,
			prometheus.GaugeValue,
			metrics.Applications[k].Processes.Running,
		)

		processStarting := prometheus.MustNewConstMetric(
			v.ProcessStarting,
			prometheus.GaugeValue,
			metrics.Applications[k].Processes.Starting,
		)

		processIdle := prometheus.MustNewConstMetric(
			v.ProcessIdle,
			prometheus.GaugeValue,
			metrics.Applications[k].Processes.Idle,
		)

		requestsActive := prometheus.MustNewConstMetric(
			v.RequestsActive,
			prometheus.GaugeValue,
			metrics.Applications[k].Requests.Active,
		)

		ch <- processRunning
		ch <- processStarting
		ch <- processIdle
		ch <- requestsActive

	}
}

func main() {
	metrics, err := controlApi.GetStatus()
	if err != nil {
		panic(err)
	}

	collector := newUnitCollector(metrics)
	collector.register()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":"+serverPort, nil))
}
