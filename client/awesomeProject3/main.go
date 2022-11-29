package main

import (
	file "awesomeProject3/protoc"
	"bufio"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"io"
	"log"
	"os"
)

func main() {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.Dial("127.0.0.1:5300", opts...)

	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}

	defer conn.Close()

	client := file.NewFileServiceClient(conn)
	stream, err := client.UploadFile(context.Background())
	if err != nil {
		panic(err)
	}
	f, err := os.Open("credentials.json")
	if err != nil {
		panic(err)
	}

	req := &file.UploadFileInfoRequest{
		Data: &file.UploadFileInfoRequest_Info{
			Info: &file.FileInfo{
				Name:        "credentials",
				Extension:   "json",
				Bucket:      "test",
				ContentType: "application/json",
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(f)
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk to buffer: ", err)
		}

		req := &file.UploadFileInfoRequest{
			Data: &file.UploadFileInfoRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to server: ", err)
		}
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	log.Printf("image uploaded with %s", res.Url)
}
