// Package kvslib provides an API which is a wrapper around RPC calls to the
// frontend.
package kvslib

import (
	"context"
	"errors"
	"log"
	"time"

	"net/rpc"

	"github.com/kttkkkng/workload/pb"
	"google.golang.org/grpc"
)

type KvslibBegin struct {
	ClientId string
}

type KvslibPut struct {
	ClientId string
	OpId     uint32
	Key      string
	Value    string
	Delay    int
}

type KvslibGet struct {
	ClientId string
	OpId     uint32
	Key      string
}

type KvslibPutResult struct {
	OpId uint32
	Err  bool
}

type KvslibGetResult struct {
	OpId  uint32
	Key   string
	Value *string
	Err   bool
}

type KvslibComplete struct {
	ClientId string
}

// NotifyChannel is used for notifying the client about a mining result.
type NotifyChannel chan ResultStruct

type ResultStruct struct {
	OpId        uint32
	StorageFail bool
	Result      *string
}

type KVS struct {
	notifyCh  NotifyChannel
	rpcClient *rpc.Client
	OpId      uint32
	ClientId  KvslibBegin
	// Add more KVS instance state here.
	grpcClientConn *grpc.ClientConn
}

func NewKVS() *KVS {
	return &KVS{
		notifyCh: nil,
	}
}

// Initialize Initializes the instance of KVS to use for connecting to the frontend,
// and the frontends IP:port. The returned notify-channel channel must
// have capacity ChCapacity and must be used by kvslib to deliver all solution
// notifications. If there is an issue with connecting, this should return
// an appropriate err value, otherwise err should be set to nil.
func (d *KVS) Initialize(clientId string, frontEndAddr string, chCapacity uint) (NotifyChannel, error) {
	d.OpId = 0

	d.ClientId = KvslibBegin{clientId}

	notifyLocal := make(chan ResultStruct, chCapacity)
	d.notifyCh = notifyLocal

	conn, err := grpc.Dial(frontEndAddr, grpc.WithInsecure())
	if err != nil {
		// log error
		log.Fatal(err)
		return nil, errors.New("cannot connect to grpc server")
	}

	// save conn to KVS struct
	d.grpcClientConn = conn

	return d.notifyCh, nil
}

// Get is a non-blocking request from the client to the system. This call is used by
// the client when it wants to get value for a key.
func (d *KVS) Get(clientId string, key string) (uint32, error) {
	d.OpId++

	client := pb.NewFrontendClient(d.grpcClientConn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	res, err := client.HandleGet(ctx, &pb.FrontendGetRequest{
		ClientId: clientId,
		OpId:     d.OpId,
		Key:      key,
	})
	if err != nil {
		return d.OpId, errors.New("HandleGet Failed")
	}

	// convert grpc to return type: ResultStruct
	reply := new(ResultStruct)
	reply.OpId = d.OpId
	reply.Result = &res.Result
	reply.StorageFail = res.StorageFail

	d.notifyCh <- *reply

	return d.OpId, nil
}

// Put is a non-blocking request from the client to the system. This call is used by
// the client when it wants to update the value of an existing key or add add a new
// key and value pair.
func (d *KVS) Put(clientId string, key string, value string, delay int) (uint32, error) {
	d.OpId++

	client := pb.NewFrontendClient(d.grpcClientConn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	res, err := client.HandlePut(ctx, &pb.FrontendPutRequest{
		ClientId: clientId,
		OpId:     d.OpId,
		Key:      key,
		Value:    value,
		Delay:    uint32(delay),
	})
	if err != nil {
		return d.OpId, errors.New("HandlePut Failed")
	}

	reply := new(ResultStruct)
	reply.OpId = d.OpId
	reply.StorageFail = res.StorageFail
	reply.Result = &res.Result
	d.notifyCh <- *reply

	return d.OpId, nil
}

// Close Stops the KVS instance from communicating with the frontend and
// from delivering any solutions via the notify-channel. If there is an issue
// with stopping, this should return an appropriate err value, otherwise err
// should be set to nil.
func (d *KVS) Close() error {
	// err := d.rpcClient.Close()
	// if err != nil {
	// 	return errors.New(err.Error())
	// }
	err := d.grpcClientConn.Close()
	if err != nil {
		return errors.New(err.Error())
	}
	return nil
}