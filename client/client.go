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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run client.go <get|set> <bucket> <key> [<value>]")
		return
	}

	command := os.Args[1]

	if (command == "get" && len(os.Args) < 4) || (command == "set" && len(os.Args) < 5) {
		switch command {
		case "get":
			fmt.Println("Usage: go run client.go get <bucket> <key>")
		case "set":
			fmt.Println("Usage: go run client.go set <bucket> <key> <value>")
		default:
			fmt.Println("Usage: go run client.go <get|set> <bucket> <key> [<value>]")
		}
		return
	}

	client, err := rpc.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Dialing:", err)
	}

	bucket := os.Args[2]
	key, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatalf("Invalid key: '%s', must be an integer.", os.Args[3])
	}

	switch command {
	case "get":
		req := GetRequest{Bucket: bucket, Key: key}
		var res GetResponse
		err := client.Call("API.Get", req, &res)
		if err != nil {
			log.Fatal("API.Get error:", err)
		}
		fmt.Printf("Value: %s\n", res.Value)

	case "set":
		value := os.Args[4]
		req := SetRequest{Bucket: bucket, Key: key, Value: value}
		var res SetResponse
		err := client.Call("API.Set", req, &res)
		if err != nil {
			log.Fatal("API.Set error:", err)
		}
		if res.OK {
			fmt.Printf("Key '%d' with value '%s' in bucket '%s' set.\n", key, value, bucket)
		} else {
			fmt.Printf("Failed to set key '%d' in bucket '%s'.\n", key, bucket)
		}
	default:
		fmt.Println("Unknown command:", command)
	}
}
