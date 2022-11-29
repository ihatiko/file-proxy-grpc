package main

import (
	"awesomeProject1/pkg/minio"
	pb "awesomeProject1/protoc"
	"bytes"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"path"
)

type Service struct {
}

const maxImageSize = 1 << 20

func (s *Service) UploadImage(stream pb.FileService_UploadFileServer) error {
	req, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.Unknown, "cannot receive image info")
	}
	imageData := bytes.Buffer{}
	imageSize := 0

	extension := req.GetInfo().GetExtension()
	bucket := req.GetInfo().GetBucket()
	name := req.GetInfo().GetName()

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
	//TODO CONFIG public dns
	url := "https://localhost:9090"
	res := &pb.UploadFileResponse{
		Url: path.Join(url, bucket, fmt.Sprintf("%s.%s", name, extension)),
	}
	err = stream.SendAndClose(res)
	if err != nil {
		return status.Errorf(codes.Unknown, "cannot send response: %v", err)
	}

	return nil
}

func main() {
	user := "AKIAIOSFODNN7EXAMPLE"
	password := "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	url := "localhost:9000"
	m := minio.NewProvider(user, password, url, false)

	app := fiber.New()

	app.Put("/test", func(c *fiber.Ctx) error {
		file, err := c.FormFile("fileUpload")

		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}
		buffer, err := file.Open()

		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}
		defer buffer.Close()

		err = m.Connect()
		if err != nil {
			// Return status 500 and minio connection error.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}

		objectName := file.Filename
		fileBuffer := buffer
		contentType := file.Header["Content-Type"][0]
		fileSize := file.Size
		bucket := "test"
		info, err := m.Put(c.Context(), bucket, objectName, contentType, fileBuffer, fileSize)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}

		log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)

		return c.JSON(fiber.Map{
			"error": false,
			"msg":   nil,
			"info":  info,
		})
	})

	app.Listen(":3000")
}
