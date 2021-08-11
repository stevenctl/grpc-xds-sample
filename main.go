package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	creds "google.golang.org/grpc/credentials/xds"
	// even if you're not using anything from this import, it should still be included to get the side-effect
	// of installing xDS balancers and resolvers (enabling xds:/// URLs in grpc.Dial)
	"google.golang.org/grpc/xds"

	"github.com/stevenctl/grpc-xds-sample/greeter"
)

func main() {
	go func() {
		err := Serve()
		log.Fatalf("finished serving: %v", err)
	}()

	client, err := NewClient()
	if err != nil {
		log.Fatalf("failed to initialize client: %v", err)
	}

	for {
		res, err := client.Hello(context.Background(), &greeter.HelloRequest{Name: name()})
		if err != nil {
			log.Printf("request failed with error: %v", err)
		}
		log.Printf(res.GetMessage())
		time.Sleep(5 * time.Second)
	}
}

// NewClient sets up a greeter client using xDS
func NewClient() (greeter.GreeterClient, error) {
	// tell the gRPC server to let xDS configure security
	clientCreds, err := creds.NewClientCredentials(creds.ClientOptions{
		// allow falling back to insecure if no security configuration is given over xDS
		FallbackCreds: insecure.NewCredentials(),
	})
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(
		// make sure to use the fully-qualified domain name (that's what Istio can resolve)
		// this will look like "xds:///greeter.xdssample.svc.cluster.local:7070"
		xdsURL(),
		// use the credentials defined above
		grpc.WithTransportCredentials(clientCreds),
	)
	if err != nil {
		return nil, err
	}
	return greeter.NewGreeterClient(conn), nil
}

// Serve starts serving the greeter service using XDS on port 7070
func Serve() error {
	// tell the gRPC server to let xDS configure security
	serverCreds, err := creds.NewServerCredentials(creds.ServerOptions{
		// allow falling back to insecure if no security configuration is given over xDS
		FallbackCreds: insecure.NewCredentials(),
	})
	if err != nil {
		return err
	}
	server := xds.NewGRPCServer(grpc.Creds(serverCreds))
	greeter.RegisterGreeterServer(server, &greeterServer{})
	listener, err := net.Listen("tcp", "[::]:7070")
	if err != nil {
		return err
	}
	return server.Serve(listener)
}

type greeterServer struct {
	greeter.UnimplementedGreeterServer
}

func (s *greeterServer) Hello(ctx context.Context, req *greeter.HelloRequest) (*greeter.HelloResponse, error) {
	return &greeter.HelloResponse{Message: fmt.Sprintf("Hello, %s! From: %s.", req.GetName(), name())}, nil
}

func name() string {
	name, err := os.Hostname()
	if err != nil {
		name = "xDS server"
	}
	return name
}

func xdsURL() string {
	params := map[string]string{
		"SERVICE_NAME":      "greeter",
		"SERVICE_NAMESPACE": "xdssample",
		"SERVICE_PORT":      "7070",
	}
	for k := range params {
		if v := os.Getenv(k); v != "" {
			params[k] = v
		}
	}
	return fmt.Sprintf(
		"xds:///%s.%s.svc.cluster.local:%s",
		params["SERVICE_NAME"],
		params["SERVICE_NAMESPACE"],
		params["SERVICE_PORT"],
	)
}
