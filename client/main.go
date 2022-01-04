package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	pb "new/helloworld/helloworld"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defaultName = "world"
)

var (
	addr     = flag.String("addr", "localhost:50052", "the address to connect to")
	name     = flag.String("name", defaultName, "Name to greet")
	filename = flag.String("f", "", "filename")
	logger   *log.Logger
)

func loginit() {

	file := currentDir() + "/client.log"
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

func main() {
	flag.Parse()
	loginit()
	_, err := os.Stat(*filename)
	if err != nil {
		log.Fatal("file not found")
	}

	data, err := ioutil.ReadFile(*filename)

	creds, err := credentials.NewClientTLSFromFile("x509/server_cert.pem", "www.mada0.com")
	if err != nil {
		log.Fatalf("failed to load credentials: %v", err)
	}
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*600)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: path.Base(*filename), Data: data})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	if strings.Contains(r.Message, "sucess") {
		logger.Println(*addr + "revice " + *filename)
	} else {
		logger.Println(*addr + "faild revice " + *filename)
	}
	log.Printf("%s", r.GetMessage())

	r, err = c.ExCommod(ctx, &pb.HelloRequest{Name: path.Base(*filename), Data: data})
	if err != nil {
		log.Fatal("could not Ex")
	}
	logger.Println(*addr + r.Message + *filename)
	log.Printf("%s", r.GetMessage())
}
