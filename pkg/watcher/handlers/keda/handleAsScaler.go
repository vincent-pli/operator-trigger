package keda

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/vincent-pli/operator-trigger/pkg/watcher/handlers/keda/externalscaler"

	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

type KedaHandler struct {
	TargetInstance  *corev1.ObjectReference
	PollingInterval uint
	Log             logr.Logger
	EpsServer       pb.ExternalScaler_StreamIsActiveServer
}

var _ cache.ResourceEventHandler = (*KedaHandler)(nil)

func New(targetInstance *corev1.ObjectReference, pollingInterval uint, log logr.Logger) (*KedaHandler, error) {
	//create svc, scaledObject before really start

	return &KedaHandler{
		TargetInstance:  targetInstance,
		PollingInterval: pollingInterval,
		Log:             log,
	}, nil

	return nil, nil
}

func (c *KedaHandler) Start() error {
	grpcServer := grpc.NewServer()
	lis, _ := net.Listen("tcp", ":6000")
	pb.RegisterExternalScalerServer(grpcServer, &ExternalScaler{})

	fmt.Println("listenting on :6000")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func (k *KedaHandler) OnAdd(obj interface{}) {
	err = epsServer.Send(&pb.IsActiveResponse{
		Result: true,
	})
}

func (k *KedaHandler) OnUpdate(oldObj, newObj interface{}) {

}

func (k *KedaHandler) OnDelete(obj interface{}) {

}

func (k *KedaHandler) IsActive(ctx context.Context, scaledObject *pb.ScaledObjectRef) (*pb.IsActiveResponse, error) {
}

func (k *KedaHandler) GetMetricSpec(context.Context, *pb.ScaledObjectRef) (*pb.GetMetricSpecResponse, error) {
}

func (k *KedaHandler) GetMetrics(_ context.Context, metricRequest *pb.GetMetricsRequest) (*pb.GetMetricsResponse, error) {
}

func (k *KedaHandler) StreamIsActive(scaledObject *pb.ScaledObjectRef, epsServer pb.ExternalScaler_StreamIsActiveServer) error {
	e.EpsServer = epsServer

	return nil
}
