package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Site struct {
	Host        string  `json:"host" yaml:"host"`
	AWSKey      string  `json:"awsKey" yaml:"awsKey"`
	AWSSecret   string  `json:"awsSecret" yaml:"awsSecret"`
	AWSRegion   string  `json:"awsRegion" yaml:"awsRegion"`
	AWSBucket   string  `json:"awsBucket" yaml:"awsBucket"`
	AWSEndpoint string  `json:"awsEndpoint,omitempty" yaml:"awsEndpoint,omitempty"`
	Users       []User  `json:"users" yaml:"users"`
	Options     Options `json:"options" yaml:"options"`
}

type User struct {
	Name     string `json:"name" yaml:"name"`
	Password string `json:"password" yaml:"password"`
}

type Options struct {
	CORS     bool   `json:"cors" yaml:"cors"`
	Gzip     bool   `json:"gzip" yaml:"gzip"`
	Website  bool   `json:"website" yaml:"website"`
	Prefix   string `json:"prefix" yaml:"prefix"`
	ForceSSL bool   `json:"forceSsl" yaml:"forceSsl"`
	Proxied  bool   `json:"proxied" yaml:"proxied"`
}

func main() {
	handler, err := ConfiguredProxyHandler()
	if err != nil {
		fmt.Printf("fatal: %v\n", err)
		return
	}

	port := flag.Int("port", 8080, "Port to listen on")

	flag.Parse()

	portStr := strconv.FormatInt(int64(*port), 10)

	log.Println("s3-proxy is listening on port " + portStr)
	log.Fatal(http.ListenAndServe(":"+portStr, handler))
}
