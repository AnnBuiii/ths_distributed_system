package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"

	"github.com/marcelloh/fastdb"
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

type API struct {
	db *fastdb.DB
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
	err := a.db.Set(req.Bucket, req.Key, []byte(req.Value))
	if err != nil {
		res.OK = false
		return err
	}
	res.OK = true
	return nil
}

func main() {
	store, err := fastdb.Open(":memory:", 100)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err = store.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	api := &API{db: store}
	rpc.Register(api)

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Listener error:", err)
	}
	fmt.Println("Server listening on :8080")
	rpc.Accept(listener)
}
