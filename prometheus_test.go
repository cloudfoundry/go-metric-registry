package metrics_test

import (
	metrics "code.cloudfoundry.org/go-metric-registry"
	"code.cloudfoundry.org/tlsconfig/certtest"
	"crypto/tls"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"log"
	"net/http"
)

var _ = Describe("PrometheusMetrics", func() {
	var (
		l = log.New(GinkgoWriter, "", log.LstdFlags)
	)

	BeforeEach(func() {
		// This is needed because the prom registry will register
		// the /metrics route with the default http mux which is
		// global
		http.DefaultServeMux = new(http.ServeMux)
	})

	It("serves metrics on a prometheus endpoint", func() {
		r := metrics.NewRegistry(l, metrics.WithServer(0))

		c := r.NewCounter("test_counter", "a counter help text for test_counter", metrics.WithMetricLabels(map[string]string{"foo": "bar"}), )

		g := r.NewGauge("test_gauge", "a gauge help text for test_gauge", metrics.WithMetricLabels(map[string]string{"bar": "baz"}), )

		c.Add(10)
		g.Set(10)
		g.Add(1)

		Eventually(func() string { return getMetrics(r.Port()) }).Should(ContainSubstring(`test_counter{foo="bar"} 10`))
		Eventually(func() string { return getMetrics(r.Port()) }).Should(ContainSubstring("a counter help text for test_counter"))
		Eventually(func() string { return getMetrics(r.Port()) }).Should(ContainSubstring(`test_gauge{bar="baz"} 11`))
		Eventually(func() string { return getMetrics(r.Port()) }).Should(ContainSubstring("a gauge help text for test_gauge"))
	})

	It("returns the metric when duplicate is created", func() {
		r := metrics.NewRegistry(l, metrics.WithServer(0))

		c := r.NewCounter("test_counter", "help text goes here")
		c2 := r.NewCounter("test_counter", "help text goes here")

		c.Add(1)
		c2.Add(2)

		Eventually(func() string {
			return getMetrics(r.Port())
		}).Should(ContainSubstring(`test_counter 3`))

		g := r.NewGauge("test_gauge", "help text goes here")
		g2 := r.NewGauge("test_gauge", "help text goes here")

		g.Add(1)
		g2.Add(2)

		Eventually(func() string {
			return getMetrics(r.Port())
		}).Should(ContainSubstring(`test_gauge 3`))
	})

	It("panics if the metric is invalid", func() {
		r := metrics.NewRegistry(l)

		Expect(func() {
			r.NewCounter("test-counter", "help text goes here")
		}).To(Panic())

		Expect(func() {
			r.NewGauge("test-counter", "help text goes here")
		}).To(Panic())
	})

	Context("WithTLSServer", func() {
		It("starts a TLS server", func() {
			ca, caFile := generateCA("someCA")
			certFile, keyFile := generateCertKeyPair(ca, "server")

			r := metrics.NewRegistry(
				l,
				metrics.WithTLSServer(0, certFile, keyFile, caFile),
			)

			g := r.NewGauge("test_gauge", "a gauge help text for test_gauge", metrics.WithMetricLabels(map[string]string{"bar": "baz"}), )
			g.Set(10)

			Eventually(func() string {
				return getMetricsTLS(r.Port(), ca)
			}).Should(ContainSubstring(`test_gauge{bar="baz"} 10`))

			addr := fmt.Sprintf("http://127.0.0.1:%s/metrics", r.Port())
			resp, err := http.Get(addr)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})
})

func getMetrics(port string) string {
	addr := fmt.Sprintf("http://127.0.0.1:%s/metrics", port)
	resp, err := http.Get(addr)
	if err != nil {
		return ""
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	Expect(err).ToNot(HaveOccurred())

	return string(respBytes)
}

func getMetricsTLS(port string, ca *certtest.Authority) string {
	caPool, err := ca.CertPool()
	if err != nil {
		log.Fatal(err)
	}

	cert, err := ca.BuildSignedCertificate("client")
	if err != nil {
		log.Fatal(err)
	}

	tlsCert, err := cert.TLSCertificate()
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{tlsCert},
				RootCAs:      caPool,
			},
		},
	}

	addr := fmt.Sprintf("https://127.0.0.1:%s/metrics", port)
	resp, err := client.Get(addr)
	if err != nil {
		return ""
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	Expect(err).ToNot(HaveOccurred())

	return string(respBytes)
}

func generateCA(caName string) (*certtest.Authority, string) {
	ca, err := certtest.BuildCA(caName)
	if err != nil {
		log.Fatal(err)
	}

	caBytes, err := ca.CertificatePEM()
	if err != nil {
		log.Fatal(err)
	}

	fileName := tmpFile(caName+".crt", caBytes)

	return ca, fileName
}

func tmpFile(prefix string, caBytes []byte) string {
	file, err := ioutil.TempFile("", prefix)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = file.Write(caBytes)
	if err != nil {
		log.Fatal(err)
	}

	return file.Name()
}

func generateCertKeyPair(ca *certtest.Authority, commonName string) (string, string) {
	cert, err := ca.BuildSignedCertificate(commonName, certtest.WithDomains(commonName))
	if err != nil {
		log.Fatal(err)
	}

	certBytes, keyBytes, err := cert.CertificatePEMAndPrivateKey()
	if err != nil {
		log.Fatal(err)
	}

	certFile := tmpFile(commonName+".crt", certBytes)
	keyFile := tmpFile(commonName+".key", keyBytes)

	return certFile, keyFile
}
