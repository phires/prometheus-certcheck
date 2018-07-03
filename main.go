package main

import (
	"bufio"
	"flag"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/phires/prometheus-certcheck/certificate"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type config struct {
	telemetryPath   string
	webAddress      string
	hostsFile       string
	refreshInterval int64
}

var (
	expiryHours = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "certificate_expiry_hours",
			Help: "Number of hours until certificate is expiring.",
		},
		[]string{"certificate"},
	)

	expiryDays = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "certificate_expiry_days",
			Help: "Number of days until certificate is expiring.",
		},
		[]string{"certificate"},
	)
)

func parseFlags() *config {
	cfg := &config{}
	flag.StringVar(&cfg.telemetryPath, "web.telemetry-path", "/metrics", "Address to listen on for web endpoints.")
	flag.StringVar(&cfg.webAddress, "web.address", ":8080", "Network address to listen on for web endpoints.")
	flag.StringVar(&cfg.hostsFile, "hosts", "", "Path to file with hostnames (one per line, format <hostname>:<port>)")
	flag.Int64Var(&cfg.refreshInterval, "interval", 60, "Path to config file.")
	flag.Parse()
	return cfg
}

func init() {
	prometheus.MustRegister(expiryDays)
	prometheus.MustRegister(expiryHours)
}

func schedule(what func(), delay time.Duration) chan bool {
	stop := make(chan bool)

	go func() {
		for {
			what()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()

	return stop
}

func main() {
	cfg := parseFlags()

	if len(cfg.hostsFile) == 0 {
		log.Fatal("No hostfile given")
	}

	http.Handle(cfg.telemetryPath, prometheus.Handler())

	evaluate := func() {
		file, err := os.Open(cfg.hostsFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scan := bufio.NewScanner(file)
		for scan.Scan() {
			result, err := certificate.GetCertificatesOfHost(scan.Text())
			if err != nil {
				log.Error(err)
			}

			for _, cert := range result.Certs {
				expireIn := float64(cert.NotAfter.Sub(time.Now()).Hours())

				expiryHours.With(prometheus.Labels{"certificate": string(cert.Subject.CommonName)}).Set(math.Floor(expireIn))
				expiryDays.With(prometheus.Labels{"certificate": string(cert.Subject.CommonName)}).Set(math.Floor(expireIn / 24))
				log.Infof("Cert: \"%s\" expire date: \"%v\" (in %f hours)", cert.Subject, cert.NotAfter, expireIn)
			}
		}
		if err := scan.Err(); err != nil {
			log.Error(err)
		}
	}

	timer := schedule(evaluate, time.Duration(cfg.refreshInterval)*time.Minute)

	if err := http.ListenAndServe(cfg.webAddress, nil); err != nil {
		panic(err)
	}

	timer <- true

}
