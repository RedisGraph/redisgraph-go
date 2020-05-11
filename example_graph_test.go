package redisgraph_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/RedisGraph/redisgraph-go"
	"github.com/gomodule/redigo/redis"
	"io/ioutil"
	"log"
	"os"
)

func ExampleGraphNew() {
	conn, _ := redis.Dial("tcp", "0.0.0.0:6379")

	graph := redisgraph.GraphNew("social", conn)

	q := "CREATE (w:WorkPlace {name:'RedisLabs'})"
	res, _ := graph.Query(q)

	q = "MATCH (w:WorkPlace) RETURN w"
	res, _ = graph.Query(q)

	res.Next()
	r := res.Record()
	w := r.GetByIndex(0).(*redisgraph.Node)
	fmt.Println(w.Label)
	// Output: WorkPlace
}

func ExampleGraphNew_pool() {
	host := "localhost:6379"
	password := ""
	pool := &redis.Pool{Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", host, redis.DialPassword(password))
	}}

	graph := redisgraph.GraphNew("social", pool.Get())

	q := "CREATE (w:WorkPlace {name:'RedisLabs'})"
	res, _ := graph.Query(q)

	q = "MATCH (w:WorkPlace) RETURN w"
	res, _ = graph.Query(q)

	res.Next()
	r := res.Record()
	w := r.GetByIndex(0).(*redisgraph.Node)
	fmt.Println(w.Label)
	// Output: WorkPlace

}

func ExampleGraphNew_tls() {
	// Consider the following helper methods that provide us with the connection details (host and password)
	// and the paths for:
	//     tls_cert - A a X.509 certificate to use for authenticating the  server to connected clients, masters or cluster peers. The file should be PEM formatted
	//     tls_key - A a X.509 private key to use for authenticating the  server to connected clients, masters or cluster peers. The file should be PEM formatted
	//	   tls_cacert - A PEM encoded CA's certificate file
	host, password := getConnectionDetails()
	tlsready, tls_cert, tls_key, tls_cacert := getTLSdetails()

	// Skip if we dont have all files to properly connect
	if tlsready == false {
		return
	}

	// Load client cert
	cert, err := tls.LoadX509KeyPair(tls_cert, tls_key)
	if err != nil {
		log.Fatal(err)
	}

	// Load CA cert
	caCert, err := ioutil.ReadFile(tls_cacert)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	clientTLSConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	// InsecureSkipVerify controls whether a client verifies the
	// server's certificate chain and host name.
	// If InsecureSkipVerify is true, TLS accepts any certificate
	// presented by the server and any host name in that certificate.
	// In this mode, TLS is susceptible to man-in-the-middle attacks.
	// This should be used only for testing.
	clientTLSConfig.InsecureSkipVerify = true

	pool := &redis.Pool{Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", host,
			redis.DialPassword(password),
			redis.DialTLSConfig(clientTLSConfig),
			redis.DialUseTLS(true),
			redis.DialTLSSkipVerify(true),
		)
	}}

	graph := redisgraph.GraphNew("social", pool.Get())

	q := "CREATE (w:WorkPlace {name:'RedisLabs'})"
	res, err := graph.Query(q)

	q = "MATCH (w:WorkPlace) RETURN w"
	res, err = graph.Query(q)

	res.Next()
	r := res.Record()
	w := r.GetByIndex(0).(*redisgraph.Node)
	fmt.Println(w.Label)
}

func getConnectionDetails() (host string, password string) {
	value, exists := os.LookupEnv("REDISGRAPH_TEST_HOST")
	host = "localhost:6379"
	password = ""
	valuePassword, existsPassword := os.LookupEnv("REDISGRAPH_TEST_PASSWORD")
	if exists && value != "" {
		host = value
	}
	if existsPassword && valuePassword != "" {
		password = valuePassword
	}
	return
}

func getTLSdetails() (tlsready bool, tls_cert string, tls_key string, tls_cacert string) {
	tlsready = false
	value, exists := os.LookupEnv("TLS_CERT")
	if exists && value != "" {
		info, err := os.Stat(value)
		if os.IsNotExist(err) || info.IsDir() {
			return
		}
		tls_cert = value
	} else {
		return
	}
	value, exists = os.LookupEnv("TLS_KEY")
	if exists && value != "" {
		info, err := os.Stat(value)
		if os.IsNotExist(err) || info.IsDir() {
			return
		}
		tls_key = value
	} else {
		return
	}
	value, exists = os.LookupEnv("TLS_CACERT")
	if exists && value != "" {
		info, err := os.Stat(value)
		if os.IsNotExist(err) || info.IsDir() {
			return
		}
		tls_cacert = value
	} else {
		return
	}
	tlsready = true
	return
}
