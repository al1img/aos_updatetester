package grpcserver

import (
	"fmt"
	"net"
	"sync"

	pb "gitpct.epam.com/epmd-aepr/aos_common/api/updatemanager"
	"google.golang.org/grpc"
)

/*******************************************************************************
 * Consts
 ******************************************************************************/

/*******************************************************************************
 * Types
 ******************************************************************************/

// Instance server instance
type Instance struct {
	sync.Mutex
	consumer   Consumer
	grpcServer *grpc.Server
	streams    map[string]pb.UpdateController_RegisterUMServer
}

// Consumer server consumer interface
type Consumer interface {
	Registered()
	Disconnected(id string, err error)
	Status(id, state, err string)
}

/*******************************************************************************
 * Public
 ******************************************************************************/

// New creates new grpc server instance
func New(url string, consumer Consumer) (instance *Instance, err error) {
	instance = &Instance{consumer: consumer, streams: make(map[string]pb.UpdateController_RegisterUMServer)}

	listener, err := net.Listen("tcp", url)
	if err != nil {
		return nil, err
	}

	instance.grpcServer = grpc.NewServer()

	pb.RegisterUpdateControllerServer(instance.grpcServer, instance)

	go instance.grpcServer.Serve(listener)

	return instance, nil
}

// RegisterUM register UM callback
func (instance *Instance) RegisterUM(stream pb.UpdateController_RegisterUMServer) (err error) {
	id := ""

	instance.consumer.Registered()

	for {
		status, err := stream.Recv()
		if err != nil {
			instance.consumer.Disconnected(id, err)

			return err
		}

		id = status.UmId

		instance.Lock()
		instance.streams[id] = stream
		instance.Unlock()

		instance.consumer.Status(status.UmId, status.UmState.String(), status.Error)
	}
}

// Close closes grpc server
func (instance *Instance) Close() (err error) {
	if instance.grpcServer != nil {
		instance.grpcServer.Stop()
	}

	return nil
}

// PrepareUpdate sends prepare update request
func (instance *Instance) PrepareUpdate(id, url string, version uint64) (err error) {
	stream, ok := instance.streams[id]
	if !ok {
		return fmt.Errorf("client %s not found", id)
	}

	if err = stream.Send(&pb.SmMessages{SmMessage: &pb.SmMessages_PrepareUpdate{
		PrepareUpdate: &pb.PrepareUpdate{Url: url, Version: version}}}); err != nil {
		return err
	}

	return nil
}

// StartUpdate sends start update request
func (instance *Instance) StartUpdate(id string) (err error) {
	stream, ok := instance.streams[id]
	if !ok {
		return fmt.Errorf("client %s not found", id)
	}

	if err = stream.Send(&pb.SmMessages{SmMessage: &pb.SmMessages_StartUpdate{}}); err != nil {
		return err
	}

	return nil
}

// ApplyUpdate sends start apply request
func (instance *Instance) ApplyUpdate(id string) (err error) {
	stream, ok := instance.streams[id]
	if !ok {
		return fmt.Errorf("client %s not found", id)
	}

	if err = stream.Send(&pb.SmMessages{SmMessage: &pb.SmMessages_ApplyUpdate{}}); err != nil {
		return err
	}

	return nil
}

// RevertUpdate sends revert apply request
func (instance *Instance) RevertUpdate(id string) (err error) {
	stream, ok := instance.streams[id]
	if !ok {
		return fmt.Errorf("client %s not found", id)
	}

	if err = stream.Send(&pb.SmMessages{SmMessage: &pb.SmMessages_RevertUpdate{}}); err != nil {
		return err
	}

	return nil
}
