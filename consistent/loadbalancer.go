package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
)

type GetRequest struct {
	Bucket string
	Key    int
}

type GetResponse struct {
	Value string
}

type SetRequest struct {
	Bucket string
	Key    int
	Value  string
}

type SetResponse struct {
	OK bool
}

// LoadBalancer quản lý các node sử dụng consistent hashing
type LoadBalancer struct {
	consistent *Consistent
	clients    map[string]*rpc.Client
	mu         sync.RWMutex
}

// NewLoadBalancer tạo một load balancer mới
func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		consistent: New(),
		clients:    make(map[string]*rpc.Client),
	}
}

// AddNode thêm một node mới vào load balancer
func (lb *LoadBalancer) AddNode(ip string) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Kết nối tới node
	client, err := rpc.Dial("tcp", ip)
	if err != nil {
		return fmt.Errorf("failed to connect to node %s: %v", ip, err)
	}

	// Thêm vào consistent hash ring
	lb.consistent.Add(ip)

	// Lưu client connection
	lb.clients[ip] = client
	log.Printf("Node %s added successfully", ip)
	return nil
}

// RemoveNode xóa một node khỏi load balancer
func (lb *LoadBalancer) RemoveNode(ip string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.consistent.Remove(ip)
	if client, ok := lb.clients[ip]; ok {
		client.Close()
		delete(lb.clients, ip)
	}
}

// GetNode lấy node phù hợp cho một key
func (lb *LoadBalancer) GetNode(key string) (string, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	return lb.consistent.Get(key)
}

// GetNodes lấy tất cả các node hiện tại
func (lb *LoadBalancer) GetNodes() []string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	return lb.consistent.Members()
}

// GetNodeCount trả về số lượng node
func (lb *LoadBalancer) GetNodeCount() int {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	return len(lb.clients)
}

// Get xử lý request Get và forward đến node phù hợp
func (lb *LoadBalancer) Get(req GetRequest, res *GetResponse) error {
	node, err := lb.GetNode(fmt.Sprintf("%s:%d", req.Bucket, req.Key))
	if err != nil {
		return errors.New("no nodes available")
	}

	lb.mu.RLock()
	client, ok := lb.clients[node]
	lb.mu.RUnlock()

	if !ok {
		return fmt.Errorf("node %s not found", node)
	}

	return client.Call("API.Get", req, res)
}

// Set xử lý request Set và forward đến node phù hợp
func (lb *LoadBalancer) Set(req SetRequest, res *SetResponse) error {
	node, err := lb.GetNode(fmt.Sprintf("%s:%d", req.Bucket, req.Key))
	if err != nil {
		return errors.New("no nodes available")
	}

	lb.mu.RLock()
	client, ok := lb.clients[node]
	lb.mu.RUnlock()

	if !ok {
		return fmt.Errorf("node %s not found", node)
	}

	return client.Call("API.Set", req, res)
}

func main() {
	port := flag.String("port", "8000", "Port to listen on (default: 8000 for load balancer)")
	flag.Parse()

	lb := NewLoadBalancer()

	// Đăng ký RPC
	rpc.Register(lb)

	addr := ":" + *port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Listener error:", err)
	}
	fmt.Printf("Load Balancer listening on %s\n", addr)

	// Chấp nhận kết nối
	rpc.Accept(listener)
}
