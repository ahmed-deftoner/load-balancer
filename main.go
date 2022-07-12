package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

type Server interface {
	Address() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request)
}

type loadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

func newLoadBalancer(port string, servers []Server) *loadBalancer {
	return &loadBalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
	}
}

func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	handleErr(err)
	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func handleErr(err error) {
	if err != nil {
		fmt.Println("error : %v", err)
		os.Exit(1)
	}
}

func main() {
	servers := []Server{
		newSimpleServer("https://www.twitter.com"),
		newSimpleServer("https://www.reddit.com"),
		newSimpleServer("https://www.bing.com"),
		newSimpleServer("https://www.amazon.com"),
	}
	lb := newLoadBalancer("8000", servers)
	handleRedirect := func(rw http.ResponseWriter, req *http.Request) {
		lb.ServeProxy(rw, req)
	}
	http.HandleFunc("/", handleRedirect)

	fmt.Println("listening on 'localhost:%s'\n", lb.port)

	http.ListenAndServe(":"+lb.port, nil)
}
