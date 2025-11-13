package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Site struct {
	Host      string  `json:"host" yaml:"host"`
	AWSKey    string  `json:"awsKey" yaml:"awsKey"`
	AWSSecret string  `json:"awsSecret" yaml:"awsSecret"`
	AWSRegion string  `json:"awsRegion" yaml:"awsRegion"`
	AWSBucket string  `json:"awsBucket" yaml:"awsBucket"`
	AWSEndpoint string `json:"awsEndpoint,omitempty" yaml:"awsEndpoint,omitempty"`
	Users     []User  `json:"users" yaml:"users"`
	Options   Options `json:"options" yaml:"options"`
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
	configFile := flag.String("config-file", "", "Path to YAML config file (enables hot-reload)")
	port := flag.Int("port", 8080, "Port to listen on")

	flag.Parse()

	var handler http.Handler
	var err error

	if *configFile != "" {
		// Use file-based configuration with hot-reload
		log.Println("Loading configuration from file: " + *configFile)
		handler, err = NewReloadableHandler(*configFile)
		if err != nil {
			fmt.Printf("fatal: %v\n", err)
			return
		}
		log.Println("Hot-reload enabled - configuration will reload automatically on file changes")
	} else {
		// Use environment variable configuration (legacy)
		handler, err = ConfiguredProxyHandler()
		if err != nil {
			fmt.Printf("fatal: %v\n", err)
			return
		}
	}

	portStr := strconv.FormatInt(int64(*port), 10)

	log.Println("s3-proxy is listening on port " + portStr)
	log.Fatal(http.ListenAndServe(":"+portStr, handler))
}
