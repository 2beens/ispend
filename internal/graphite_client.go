package internal

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// GraphiteClient is a struct that defines the relevant properties of a graphite connection
type GraphiteClient struct {
	Host       string
	Port       int
	Protocol   string
	Timeout    time.Duration
	Prefix     string
	conn       net.Conn
	nop        bool
	DisableLog bool
}

// defaultTimeout is the default number of seconds that we're willing to wait
// before forcing the connection establishment to fail
const defaultTimeout = 5

// IsNop is a getter for *graphite.GraphiteClient.nop
func (gc *GraphiteClient) IsNop() bool {
	return gc.nop
}

// Given a GraphiteClient struct, Connect populates the GraphiteClient.conn field with an
// appropriate TCP connection
func (gc *GraphiteClient) Connect() error {
	if gc.IsNop() {
		return nil
	}

	if gc.conn != nil {
		err := gc.conn.Close()
		if err != nil {
			log.Printf("gc client Connect(): failed to close connection: %s", err.Error())
		}
	}

	address := fmt.Sprintf("%s:%d", gc.Host, gc.Port)

	if gc.Timeout == 0 {
		gc.Timeout = defaultTimeout * time.Second
	}

	var err error
	var conn net.Conn

	if gc.Protocol == "udp" {
		udpAddr, err := net.ResolveUDPAddr("udp", address)
		if err != nil {
			return err
		}
		conn, err = net.DialUDP(gc.Protocol, nil, udpAddr)
		if err != nil {
			log.Errorf("error 11098: %s", err)
		}
	} else {
		conn, err = net.DialTimeout(gc.Protocol, address, gc.Timeout)
	}

	if err != nil {
		return err
	}

	gc.conn = conn
	return nil
}

// Given a GraphiteClient struct, Disconnect closes the GraphiteClient.conn field
func (gc *GraphiteClient) Disconnect() error {
	err := gc.conn.Close()
	gc.conn = nil
	return err
}

// Given a Metric struct, the SendMetric method sends the supplied metric to the
// GraphiteClient connection that the method is called upon
func (gc *GraphiteClient) SendMetric(metric Metric) error {
	metrics := make([]Metric, 1)
	metrics[0] = metric

	return gc.sendMetrics(metrics)
}

// Given a slice of Metrics, the SendMetrics method sends the metrics, as a
// batch, to the GraphiteClient connection that the method is called upon
func (gc *GraphiteClient) SendMetrics(metrics []Metric) error {
	return gc.sendMetrics(metrics)
}

// sendMetrics is an internal function that is used to write to the TCP
// connection in order to communicate metrics to the remote GraphiteClient host
func (gc *GraphiteClient) sendMetrics(metrics []Metric) error {
	if gc.IsNop() {
		if !gc.DisableLog {
			for _, metric := range metrics {
				log.Printf("GraphiteClient: %s\n", metric)
			}
		}
		return nil
	}

	zeroedMetric := Metric{} // ignore uninitialized metrics
	buf := bytes.NewBufferString("")
	for _, metric := range metrics {
		if metric == zeroedMetric {
			continue // ignore uninitialized metrics
		}
		if metric.Timestamp == 0 {
			metric.Timestamp = time.Now().Unix()
		}
		metricName := ""
		if gc.Prefix != "" {
			metricName = fmt.Sprintf("%s.%s", gc.Prefix, metric.Name)
		} else {
			metricName = metric.Name
		}
		if gc.Protocol == "udp" {
			_, err := fmt.Fprintf(gc.conn, "%s %s %d\n", metricName, metric.Value, metric.Timestamp)
			if err != nil {
				log.Printf("gc client - failed to send metric [%s]: %s", metricName, err.Error())
			}
			continue
		}
		buf.WriteString(fmt.Sprintf("%s %s %d\n", metricName, metric.Value, metric.Timestamp))
	}
	if gc.Protocol == "tcp" {
		_, err := gc.conn.Write(buf.Bytes())
		//fmt.Print("Sent msg:", buf.String(), "'")
		if err != nil {
			return err
		}
	}
	return nil
}

// The SimpleSend method can be used to just pass a metric name and value and
// have it be sent to the GraphiteClient host with the current timestamp
func (gc *GraphiteClient) SimpleSend(stat string, value string) bool {
	metrics := make([]Metric, 1)
	metrics[0] = NewMetric(stat, value, time.Now().Unix())
	err := gc.sendMetrics(metrics)
	if err != nil {
		log.Errorf("graphite client failed to (simple) send metric [%s]: %s", stat, err.Error())
		return false
	}
	return true
}

func (gc *GraphiteClient) SimpleSendInt(stat string, value int) bool {
	valueStr := strconv.Itoa(value)
	return gc.SimpleSend(stat, valueStr)
}

// NewGraphite is a factory method that's used to create a new GraphiteClient
func NewGraphite(host string, port int) (*GraphiteClient, error) {
	return GraphiteFactory("tcp", host, port, "")
}

// NewGraphiteWithMetricPrefix is a factory method that's used to create a new GraphiteClient with a metric prefix
func NewGraphiteWithMetricPrefix(host string, port int, prefix string) (*GraphiteClient, error) {
	return GraphiteFactory("tcp", host, port, prefix)
}

// When a UDP connection to GraphiteClient is required
func NewGraphiteUDP(host string, port int) (*GraphiteClient, error) {
	return GraphiteFactory("udp", host, port, "")
}

// NewGraphiteNop is a factory method that returns a GraphiteClient struct but will
// not actually try to send any packets to a remote host and, instead, will just
// log. This is useful if you want to use GraphiteClient in a project but don't want
// to make GraphiteClient a requirement for the project.
func NewGraphiteNop(host string, port int) *GraphiteClient {
	graphiteNop, _ := GraphiteFactory("nop", host, port, "")
	return graphiteNop
}

func GraphiteFactory(protocol string, host string, port int, prefix string) (*GraphiteClient, error) {
	var graphite *GraphiteClient

	switch protocol {
	case "tcp":
		graphite = &GraphiteClient{Host: host, Port: port, Protocol: "tcp", Prefix: prefix}
	case "udp":
		graphite = &GraphiteClient{Host: host, Port: port, Protocol: "udp", Prefix: prefix}
	case "nop":
		graphite = &GraphiteClient{Host: host, Port: port, nop: true}
	default:
		return nil, errors.New("graphite client error: unknown protocol")
	}

	err := graphite.Connect()
	if err != nil {
		return nil, err
	}

	return graphite, nil
}
