package main

import (
	"awesomeProject1/pkg/minio"
	"awesomeProject1/protoc"
	"bytes"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net"
	"path"
)

type Service struct {
}

const maxImageSize = 1 << 20

func (s *Service) UploadFile(stream file.FileService_UploadFileServer) error {
	req, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.Unknown, "cannot receive image info")
	}
	imageData := bytes.Buffer{}
	imageSize := 0

	extension := req.GetInfo().GetExtension()
	bucket := req.GetInfo().GetBucket()
	name := req.GetInfo().GetName()
	contentType := req.GetInfo().GetContentType()

	for {
		log.Print("waiting to receive more data")
		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err)
		}

		chunk := req.GetChunkData()
		size := len(chunk)

		log.Printf("received a chunk with size: %d", size)

		imageSize += size
		if imageSize > maxImageSize {
			return status.Errorf(codes.InvalidArgument, "image is too large: %d > %d", imageSize, maxImageSize)
		}
		_, err = imageData.Write(chunk)
		if err != nil {
			return status.Errorf(codes.Internal, "cannot write chunk data: %v", err)
		}
	}

	user := "AKIAIOSFODNN7EXAMPLE"
	password := "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	url := "localhost:9000"
	m, err := minio.NewProvider(user, password, url, false)
	if err != nil {
		return status.Errorf(codes.Internal, "cannot open minio connect: %v", err)
	}

	reader := bytes.NewReader(imageData.Bytes())
	_, err = m.Put(stream.Context(), bucket, fmt.Sprintf("%s.%s", name, extension), contentType, reader, int64(imageSize))

	if err != nil {
		return status.Errorf(codes.Internal, "cannot save image: %v", err)
	}

	res := &file.UploadFileResponse{
		Url: path.Join(url, bucket, fmt.Sprintf("%s.%s", name, extension)),
	}
	err = stream.SendAndClose(res)
	if err != nil {
		return status.Errorf(codes.Unknown, "cannot send response: %v", err)
	}

	return nil
}

func main() {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	listener, err := net.Listen("tcp", ":5300")
	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}
	file.RegisterFileServiceServer(grpcServer, &Service{})
	fmt.Println("Start")
	grpcServer.Serve(listener)
}
