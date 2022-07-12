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

func (s *simpleServer) Address() string {
	return s.addr
}

func (s *simpleServer) IsAlive() bool {
	return true
}

func (s *simpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	s.proxy.ServeHTTP(rw, req)
}

func (lb *loadBalancer) GetNextAvailableServer() Server {
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]
	for !server.IsAlive() {
		lb.roundRobinCount++
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
	}
	lb.roundRobinCount++
	return server
}

func (lb *loadBalancer) ServeProxy(rw http.ResponseWriter, req *http.Request) {
	target := lb.GetNextAvailableServer()
	fmt.Println("forwarding request to address %q\n", target.Address())
	target.Serve(rw, req)
}

func main() {
	servers := []Server{
		newSimpleServer("https://www.duckduckgo.com"),
		newSimpleServer("https://www.pornhub.com"),
		newSimpleServer("https://www.facebook.com"),
		newSimpleServer("https://www.google.com"),
	}
	lb := newLoadBalancer("8000", servers)
	handleRedirect := func(rw http.ResponseWriter, req *http.Request) {
		lb.ServeProxy(rw, req)
	}
	http.HandleFunc("/", handleRedirect)

	fmt.Println("listening on 'localhost:%s'\n", lb.port)

	http.ListenAndServe(":"+lb.port, nil)
}
