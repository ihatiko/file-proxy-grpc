package main

import (
	"awesomeProject1/pkg/minio"
	pb "awesomeProject1/protoc"
	"context"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"log"
)

//https://dev.to/techschoolguru/upload-file-in-chunks-with-client-streaming-grpc-golang-4loc
type FileServer struct {
}

func (s FileServer) UploadImage(ctx context.Context, opts ...grpc.CallOption) (pb.FileService_UploadImageClient, error) {
	return nil, nil
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
