package main

import (
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

// NodeRequest request cho AddNode/RemoveNode
type NodeRequest struct {
	IP string
}

// NodeResponse response cho AddNode/RemoveNode
type NodeResponse struct {
	OK    bool
	Error string
}

// GetNodeRequest request cho GetNodeId
type GetNodeRequest struct {
	IP string
}

// GetNodeResponse response cho GetNodeId
type GetNodeResponse struct {
	NodeID uint32
	Error  string
}

func printUsage(command string) {
	switch command {
	case "get":
		fmt.Println("Usage: go run client.go get <bucket> <key>")
	case "set":
		fmt.Println("Usage: go run client.go set <bucket> <key> <value>")
	case "add-node":
		fmt.Println("Usage: go run client.go add-node <ip:port>")
	case "remove-node":
		fmt.Println("Usage: go run client.go remove-node <ip:port>")
	case "get-node-id":
		fmt.Println("Usage: go run client.go get-node-id <ip:port>")
	case "list-nodes":
		fmt.Println("Usage: go run client.go list-nodes")
	default:
		fmt.Println("Usage: go run client.go <get|set|add-node|remove-node|get-node-id|list-nodes> [args...]")
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run client.go <get|set|add-node|remove-node|get-node-id|list-nodes> [args...]")
		return
	}

	command := os.Args[1]

	// Validate arguments for each command
	switch command {
	case "get":
		if len(os.Args) < 4 {
			printUsage(command)
			return
		}
	case "get-node-id":
		if len(os.Args) < 3 {
			printUsage(command)
			return
		}
	case "set":
		if len(os.Args) < 5 {
			printUsage(command)
			return
		}
	case "add-node", "remove-node":
		if len(os.Args) < 3 {
			printUsage(command)
			return
		}
	case "list-nodes":
		// No additional args needed
	default:
		printUsage("")
		return
	}

	// Connect to load balancer (port 8000)
	client, err := rpc.Dial("tcp", "localhost:8000")
	if err != nil {
		log.Fatal("Dialing:", err)
	}

	switch command {
	case "get":
		bucket := os.Args[2]
		key, err := strconv.Atoi(os.Args[3])
		if err != nil {
			log.Fatalf("Invalid key: '%s', must be an integer.", os.Args[3])
		}
		req := GetRequest{Bucket: bucket, Key: key}
		var res GetResponse
		err = client.Call("LoadBalancer.Get", req, &res)
		if err != nil {
			log.Fatal("LoadBalancer.Get error:", err)
		}
		fmt.Printf("Value: %s\n", res.Value)

	case "set":
		bucket := os.Args[2]
		key, err := strconv.Atoi(os.Args[3])
		if err != nil {
			log.Fatalf("Invalid key: '%s', must be an integer.", os.Args[3])
		}
		value := os.Args[4]
		req := SetRequest{Bucket: bucket, Key: key, Value: value}
		var res SetResponse
		err = client.Call("LoadBalancer.Set", req, &res)
		if err != nil {
			log.Fatal("LoadBalancer.Set error:", err)
		}
		if res.OK {
			fmt.Printf("Key '%d' with value '%s' in bucket '%s' set.\n", key, value, bucket)
		} else {
			fmt.Printf("Failed to set key '%d' in bucket '%s'.\n", key, bucket)
		}

	case "add-node":
		ip := os.Args[2]
		req := NodeRequest{IP: ip}
		var res NodeResponse
		err := client.Call("LoadBalancer.AddNodeRPC", req, &res)
		if err != nil {
			log.Fatal("LoadBalancer.AddNodeRPC error:", err)
		}
		if res.OK {
			fmt.Printf("Node '%s' added successfully.\n", ip)
		} else {
			fmt.Printf("Failed to add node '%s': %s\n", ip, res.Error)
		}

	case "remove-node":
		ip := os.Args[2]
		req := NodeRequest{IP: ip}
		var res NodeResponse
		err := client.Call("LoadBalancer.RemoveNodeRPC", req, &res)
		if err != nil {
			log.Fatal("LoadBalancer.RemoveNodeRPC error:", err)
		}
		if res.OK {
			fmt.Printf("Node '%s' removed successfully.\n", ip)
		} else {
			fmt.Printf("Failed to remove node '%s': %s\n", ip, res.Error)
		}

	case "get-node-id":
		ip := os.Args[2]
		req := GetNodeRequest{IP: ip}
		var res GetNodeResponse
		err := client.Call("LoadBalancer.GetNodeId", req, &res)
		if err != nil {
			log.Fatal("LoadBalancer.GetNodeId error:", err)
		}
		if res.Error != "" {
			fmt.Printf("Error: %s\n", res.Error)
		} else {
			fmt.Printf("Node ID for '%s': %d\n", ip, res.NodeID)
		}

	case "list-nodes":
		var res []string
		err := client.Call("LoadBalancer.GetNodesRPC", struct{}{}, &res)
		if err != nil {
			log.Fatal("LoadBalancer.GetNodesRPC error:", err)
		}
		if len(res) == 0 {
			fmt.Println("No nodes registered.")
		} else {
			fmt.Println("Registered nodes:")
			for _, node := range res {
				fmt.Printf("  - %s\n", node)
			}
		}

	default:
		fmt.Println("Unknown command:", command)
		printUsage("")
	}
}
