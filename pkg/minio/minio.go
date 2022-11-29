package minio

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"log"
)

type Provider struct {
	client   *minio.Client
	user     string
	password string
	url      string
	ssl      bool
}

func NewProvider(user string, password string, url string, ssl bool) *Provider {
	return &Provider{user: user, password: password, url: url, ssl: ssl}
}

func (m *Provider) Connect() error {
	var err error
	m.client, err = minio.New(m.url, &minio.Options{
		Creds:  credentials.NewStaticV4(m.user, m.password, ""),
		Secure: m.ssl,
	})
	if err != nil {
		log.Fatalln(err)
	}

	return err
}

func (m *Provider) Put(ctx context.Context, bucket, name, contentType string, buffer io.Reader, size int64) (minio.UploadInfo, error) {
	return m.client.PutObject(ctx, bucket, name, buffer, size, minio.PutObjectOptions{ContentType: contentType})
}
