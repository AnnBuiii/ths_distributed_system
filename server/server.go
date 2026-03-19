package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"time"

	"github.com/marcelloh/fastdb"
)

// RPC Request/Response types
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

func getAPIPort(nodeID string) string {
	if len(nodeID) >= 1 {
		return "600" + nodeID[len(nodeID)-1:]
	}
	return "6001"
}

type API struct {
	db   *fastdb.DB
	node *Node
}

func NewAPI(db *fastdb.DB, node *Node) *API {
	return &API{db: db, node: node}
}

func (a *API) Get(req GetRequest, res *GetResponse) error {
	value, ok := a.db.Get(req.Bucket, req.Key)
	if !ok {
		return errors.New("key not found")
	}
	res.Value = string(value)
	return nil
}

func (a *API) Set(req SetRequest, res *SetResponse) error {
	if !a.node.IsLeader() {
		return fmt.Errorf("not leader (current leader: %s)", a.node.GetLeaderID())
	}

	err := a.db.Set(req.Bucket, req.Key, []byte(req.Value))
	if err != nil {
		res.OK = false
		return err
	}

	a.node.ReplicateToPeers(req.Bucket, req.Key, []byte(req.Value))

	res.OK = true
	return nil
}

func (a *API) Delete(req DeleteRequest, res *DeleteResponse) error {
	if !a.node.IsLeader() {
		return fmt.Errorf("not leader (current leader: %s)", a.node.GetLeaderID())
	}

	a.db.Del(req.Bucket, req.Key)

	a.node.ReplicateDeleteToPeers(req.Bucket, req.Key)

	res.OK = true
	return nil
}

func (a *API) Status(_ struct{}, res *StatusResponse) error {
	res.NodeID = a.node.ID
	res.IsLeader = a.node.IsLeader()
	res.LeaderID = a.node.GetLeaderID()
	res.Peers = a.node.Peers.ToIDs()
	return nil
}

func getNodeID() string {
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		return "node-01"
	}
	return nodeID
}

func main() {
	nodeID := getNodeID()
	log.Printf("Starting node %s", nodeID)

	store, err := fastdb.Open(":memory:", 100)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	node := NewNode(nodeID, store)

	peerListener, err := node.NewListener()
	if err != nil {
		log.Fatal("Peer listener error:", err)
	}
	defer peerListener.Close()

	rpcServer := rpc.NewServer()

	rpcServer.Register(node)

	api := NewAPI(store, node)
	rpcServer.Register(api)

	go rpcServer.Accept(peerListener)
	log.Printf("Peer communication listening on %s", node.Addr)

	node.ConnectToPeers()
	log.Printf("Node %s connected to peers: %v", nodeID, node.Peers.ToIDs())

	time.Sleep(2 * time.Second)
	node.Elect()

	apiListener, err := net.Listen("tcp", ":"+getAPIPort(nodeID))
	if err != nil {
		log.Fatal("API listener error:", err)
	}
	defer apiListener.Close()

	log.Printf("API server listening on :%s", getAPIPort(nodeID))
	log.Printf("Node %s is ready. Leader: %s", nodeID, node.GetLeaderID())

	rpcServer.Accept(apiListener)
}
