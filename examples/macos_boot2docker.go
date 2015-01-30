package main

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/samalba/dockerclient"
	"io/ioutil"
	"log"
)

func main() {
	cert, err := tls.LoadX509KeyPair("/Users/username/.boot2docker/certs/boot2docker-vm/cert.pem", "/Users/username/.boot2docker/certs/boot2docker-vm/key.pem")
	if err != nil {
		log.Fatal(err)
	}

	caCert, err := ioutil.ReadFile("/Users/username/.boot2docker/certs/boot2docker-vm/ca.pem")
	if err != nil {
		log.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlc := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		RootCAs:            caCertPool,
	}

	// Replace url if your remote docker one
	docker, err := dockerclient.NewDockerClient("tcp://192.168.59.103:2376", tlc)
	if err != nil {
		log.Fatal(err)
	}

	//...
}
