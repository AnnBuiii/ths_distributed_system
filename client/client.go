package main

import (
	"errors"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strconv"
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

type DeleteRequest struct {
	Bucket string
	Key    int
}

type DeleteResponse struct {
	OK bool
}

type StatusResponse struct {
	NodeID   string
	IsLeader bool
	LeaderID string
	Peers    []string
}

var nodePorts = map[string]string{
	"node-01": "6001",
	"node-02": "6002",
	"node-03": "6003",
	"node-04": "6004",
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "get":
		handleGet()
	case "set":
		handleSet()
	case "delete", "del":
		handleDelete()
	case "status":
		handleStatus()
	case "leader":
		findLeader()
	default:
		fmt.Println("Unknown command:", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage: go run client.go <command> [args]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  get <node> <bucket> <key>           Get value from any node")
	fmt.Println("  set <node> <bucket> <key> <value>   Set value (must be leader)")
	fmt.Println("  delete <node> <bucket> <key>        Delete key (must be leader)")
	fmt.Println("  status <node>                       Show node status")
	fmt.Println("  leader                              Find current leader")
	fmt.Println("")
	fmt.Println("Nodes: node-01 (6001), node-02 (6002), node-03 (6003), node-04 (6004)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run client.go get node-01 users 1")
	fmt.Println("  go run client.go set node-01 users 1 'John Doe'")
	fmt.Println("  go run client.go status node-01")
	fmt.Println("  go run client.go leader")
}

func getClient(node string) *rpc.Client {
	port, exists := nodePorts[node]
	if !exists {
		// Try to parse as direct port
		port = node
	}

	addr := fmt.Sprintf("localhost:%s", port)
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("Cannot connect to %s: %v", addr, err)
	}
	return client
}

func handleGet() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: go run client.go get <node> <bucket> <key>")
		fmt.Println("Example: go run client.go get node-01 users 1")
		return
	}

	node := os.Args[2]
	bucket := os.Args[3]
	key, err := strconv.Atoi(os.Args[4])
	if err != nil {
		log.Fatalf("Invalid key: '%s', must be an integer.", os.Args[4])
	}

	client := getClient(node)
	defer client.Close()

	req := GetRequest{Bucket: bucket, Key: key}
	var res GetResponse
	err = client.Call("API.Get", req, &res)
	if err != nil {
		log.Fatal("API.Get error:", err)
	}
	fmt.Printf("Value: %s\n", res.Value)
}

func handleSet() {
	if len(os.Args) < 6 {
		fmt.Println("Usage: go run client.go set <node> <bucket> <key> <value>")
		fmt.Println("Example: go run client.go set node-01 users 1 'John Doe'")
		return
	}

	node := os.Args[2]
	bucket := os.Args[3]
	key, err := strconv.Atoi(os.Args[4])
	if err != nil {
		log.Fatalf("Invalid key: '%s', must be an integer.", os.Args[4])
	}
	value := os.Args[5]

	client := getClient(node)
	defer client.Close()

	req := SetRequest{Bucket: bucket, Key: key, Value: value}
	var res SetResponse
	err = client.Call("API.Set", req, &res)
	if err != nil {
		if err.Error() == "not leader" {
			fmt.Println("Error: This node is not the leader. Use 'leader' command to find leader.")
		} else {
			log.Fatal("API.Set error:", err)
		}
		return
	}
	if res.OK {
		fmt.Printf("Key '%d' with value '%s' in bucket '%s' set successfully.\n", key, value, bucket)
	} else {
		fmt.Printf("Failed to set key '%d' in bucket '%s'.\n", key, bucket)
	}
}

func handleDelete() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: go run client.go delete <node> <bucket> <key>")
		fmt.Println("Example: go run client.go delete node-01 users 1")
		return
	}

	node := os.Args[2]
	bucket := os.Args[3]
	key, err := strconv.Atoi(os.Args[4])
	if err != nil {
		log.Fatalf("Invalid key: '%s', must be an integer.", os.Args[4])
	}

	client := getClient(node)
	defer client.Close()

	req := DeleteRequest{Bucket: bucket, Key: key}
	var res DeleteResponse
	err = client.Call("API.Delete", req, &res)
	if err != nil {
		if err.Error() == "not leader" {
			fmt.Println("Error: This node is not the leader. Use 'leader' command to find leader.")
		} else {
			log.Fatal("API.Delete error:", err)
		}
		return
	}
	if res.OK {
		fmt.Printf("Key '%d' in bucket '%s' deleted successfully.\n", key, bucket)
	} else {
		fmt.Printf("Failed to delete key '%d' in bucket '%s'.\n", key, bucket)
	}
}

func handleStatus() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run client.go status <node>")
		fmt.Println("Example: go run client.go status node-01")
		return
	}

	node := os.Args[2]
	client := getClient(node)
	defer client.Close()

	var res StatusResponse
	err := client.Call("API.Status", struct{}{}, &res)
	if err != nil {
		log.Fatal("API.Status error:", err)
	}

	fmt.Printf("Node ID:   %s\n", res.NodeID)
	fmt.Printf("Is Leader: %v\n", res.IsLeader)
	fmt.Printf("Leader ID: %s\n", res.LeaderID)
	fmt.Printf("Peers:     %v\n", res.Peers)
}

func findLeader() {
	fmt.Println("Checking all nodes to find leader...")

	for node, port := range nodePorts {
		client, err := rpc.Dial("tcp", fmt.Sprintf("localhost:%s", port))
		if err != nil {
			fmt.Printf("Node %s: unreachable\n", node)
			continue
		}

		var res StatusResponse
		err = client.Call("API.Status", struct{}{}, &res)
		client.Close()

		if err != nil {
			fmt.Printf("Node %s: error getting status\n", node)
			continue
		}

		if res.IsLeader {
			fmt.Printf("\n>>> LEADER FOUND: %s (port %s)\n", node, port)
			fmt.Printf("    Use 'go run client.go set %s <bucket> <key> <value>' for writes\n", node)
			return
		}
	}

	fmt.Println("\nNo leader found. Cluster may be electing...")
}

// Helper function to check if error is "not leader"
func isNotLeaderError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, errors.New("not leader")) || err.Error() == "not leader"
}
