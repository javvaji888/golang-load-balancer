package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Server defines the behavior of proxy servers.
type Server interface {
	Address() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request)
}

// SimpleServer implements the Server interface with a reverse proxy.
type SimpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

// NewSimpleServer creates a new instance of SimpleServer.
func NewSimpleServer(addr string) *SimpleServer {
	serverURL, err := url.Parse(addr)
	if err != nil {
		log.Fatalf("Failed to parse server address: %v", err)
	}

	return &SimpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverURL),
	}
}

// Address returns the server address.
func (s *SimpleServer) Address() string {
	return s.addr
}

// IsAlive checks the health of the server. Currently, it returns true for simplicity.
func (s *SimpleServer) IsAlive() bool {
	// In a real implementation, you might want to make a health check request here.
	return true
}

// Serve proxies the request to the underlying server.
func (s *SimpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	s.proxy.ServeHTTP(rw, req)
}

// LoadBalancer manages load distribution among multiple servers.
type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

// NewLoadBalancer creates a new LoadBalancer.
func NewLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port:    port,
		servers: servers,
	}
}

// GetNextAvailableServer retrieves the next server available for handling requests.
func (lb *LoadBalancer) GetNextAvailableServer() Server {
	var server Server
	for {
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
		if server.IsAlive() {
			break
		}
		lb.roundRobinCount++
	}
	lb.roundRobinCount++
	return server
}

// ServeProxy forwards requests to the next available server.
func (lb *LoadBalancer) ServeProxy(rw http.ResponseWriter, r *http.Request) {
	targetServer := lb.GetNextAvailableServer()
	fmt.Printf("Forwarding request to address %q\n", targetServer.Address())
	targetServer.Serve(rw, r)
}

func main() {
	servers := []Server{
		NewSimpleServer("https://www.amazon.com"),
		NewSimpleServer("http://www.yahoo.com"),
		NewSimpleServer("http://www.instagram.com"),
	}
	lb := NewLoadBalancer("8000", servers)
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		lb.ServeProxy(rw, req)
	})

	log.Printf("Serving requests at 'localhost:%s'\n", lb.port)
	if err := http.ListenAndServe(":"+lb.port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
