package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vault"
	pb "vault/pb"

	ratelimitkit "github.com/go-kit/kit/ratelimit"
	"golang.org/x/time/rate"

	//"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {
	var (
		httpAddr = flag.String("http", ":8080",
			"http listen address")
		gRPCAddr = flag.String("grpc", ":8081",
			"gRPC listen address")
	)
	flag.Parse()
	//ctx := context.Background()
	srv := vault.NewService()
	errChan := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%", <-c)
	}()
	hashEndpoint := vault.MakeHashEndpoint(srv)
	{
		hashEndpoint = ratelimitkit.NewDelayingLimiter(rate.NewLimiter(rate.Every(time.Second), 5))(hashEndpoint)
	}
	validateEndpoint := vault.MakeValidateEndpoint(srv)
	{
		validateEndpoint = ratelimitkit.NewDelayingLimiter(rate.NewLimiter(rate.Every(time.Second), 5))(validateEndpoint)
	}
	endpoints := vault.Endpoints{
		HashEndpoint:     hashEndpoint,
		ValidateEndpoint: validateEndpoint,
	}
	go func() {
		log.Println("http:", *httpAddr)
		handler := vault.NewHTTPServer(endpoints)
		errChan <- http.ListenAndServe(*httpAddr, handler)
	}()
	go func() {
		listener, err := net.Listen("tcp", *gRPCAddr)
		if err != nil {
			errChan <- err
			return
		}
		log.Println("grpc:", *gRPCAddr)
		handler := vault.NewGRPCServer(endpoints)
		gRPCServer := grpc.NewServer()
		pb.RegisterVaultServer(gRPCServer, handler)
		errChan <- gRPCServer.Serve(listener)
	}()
	log.Fatalln(<-errChan)
}
