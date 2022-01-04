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
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	pb "new/helloworld/helloworld"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	port   = flag.Int("port", 50052, "The server port")
	logger *log.Logger
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

func loginit() {

	file := currentDir() + "/server.log"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	logger = log.New(logFile, "[rpc]", log.LstdFlags|log.Lshortfile|log.LUTC)
}

func currentDir() string {
	path, err := os.Executable()
	if err != nil {
		log.Printf(err.Error())
	}
	dir := filepath.Dir(path)
	return dir
}

func banlist() (slice []string) {
	data, err := ioutil.ReadFile(currentDir() + "/ban.json")
	if err != nil {
		log.Fatalln("ban.json not found")
	}

	var s []string
	err = json.Unmarshal(data, &s)
	if err != nil {
		log.Fatalln("json err")
	}
	return s
}

func contain(fi string) bool {
	fl := true
	file, err := os.Open("/tmp/" + fi)
	if err != nil {
		log.Fatalf("err")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		for _, v := range banlist() {

			if strings.Contains(line, v) {
				fl = false
			}
		}
	}
	return fl
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	err := ioutil.WriteFile("/tmp/"+in.GetName(), in.GetData(), 0655)
	if err != nil {
		return &pb.HelloReply{Message: "send " + in.GetName() + "failed"}, nil
	} else {
		logger.Println("revice " + in.GetName())
		return &pb.HelloReply{Message: "send " + in.GetName() + "sucess"}, nil
	}
}

func (s *server) ExCommod(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	if path.Ext(in.Name) == ".sh" && contain(in.GetName()) {
		command := exec.Command("/bin/sh", "-c", "/tmp/"+in.GetName())
		stdout, _ := command.StdoutPipe()
		stderr, _ := command.StderrPipe()
		err := command.Start()
		if err != nil {
			log.Printf(err.Error())
		}
		out_st, _ := ioutil.ReadAll(stdout)
		out_err, _ := ioutil.ReadAll(stderr)
		stdout.Close()

		// f, err := os.OpenFile(currentDir()+"/server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		// if err != nil {
		// 	log.Fatal("cant creat file")
		// }
		if string(out_st) != "" {
			logger.Println("result " + string(out_st))
			// f.WriteString(time.Now().Format("20060102150405") + "result: " + string(out_st) + "\n")
		}
		if string(out_err) != "" {
			logger.Println("error: " + string(out_err))
			// f.WriteString(time.Now().Format("20060102150405") + "error: " + string(out_err) + "\n")
		} else {
			logger.Println("sucess")
			// f.WriteString(time.Now().Format("20060102150405") + "success: " + "\n")
		}

		// defer f.Close()
		return &pb.HelloReply{Message: "sucess:" + in.GetName()}, nil
	} else {
		logger.Println("invalid command or not script file")
		return &pb.HelloReply{Message: "invalid command or not script file" + in.GetName()}, nil
	}

}

func main() {
	loginit()
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create an array of gRPC options with the credentials
	creds, err := credentials.NewServerTLSFromFile("x509/server_cert.pem", "x509/server_key.pem")
	if err != nil {
		log.Fatalf("failed to create credentials: %v", err)
	}

	var option = []grpc.ServerOption{
		grpc.MaxRecvMsgSize(209715200),
		grpc.MaxSendMsgSize(209715200),
		grpc.Creds(creds),
	}

	s := grpc.NewServer(option...)
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
