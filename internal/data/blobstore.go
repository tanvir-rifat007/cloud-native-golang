package data

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BlobStore is the abstraction for storing binary blobs (arbitrary files) in S3.
type BlobStore struct {
	bucket string
	Client *s3.Client
	log    *zap.Logger
}

// NewBlobStoreOptions for NewBlobStore.
type NewBlobStoreOptions struct {
	Bucket    string
	Config    aws.Config
	Log       *zap.Logger
	PathStyle bool
}

// NewBlobStore with the given options.
// If no logger is provided, logs are discarded.
func NewBlobStore(opts NewBlobStoreOptions) *BlobStore {
	if opts.Log == nil {
		opts.Log = zap.NewNop()
	}

	client := s3.NewFromConfig(opts.Config, func(o *s3.Options) {
		o.UsePathStyle = opts.PathStyle
	})

	return &BlobStore{
		bucket: opts.Bucket,
		Client: client,
		log:    opts.Log,
	}
}

func (b *BlobStore) Bucket() string {
	return b.bucket
}


// Put a blob in the bucket under key with the given contentType.
func (b *BlobStore) Put(ctx context.Context, bucket, key, contentType string, blob io.Reader) error {
	_, err := b.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucket,
		Key:         &key,
		Body:        blob,
		ContentType: &contentType,
	})
	return err
}

// Get a blob from the bucket under key.
// If there is nothing there, returns nil and no error.
func (b *BlobStore) Get(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	getObjectOutput, err := b.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if getObjectOutput == nil {
		return nil, nil
	}
	return getObjectOutput.Body, err
}

// Delete a blob from the bucket under key.
// Deleting where nothing exists does nothing and returns no error.
func (b *BlobStore) Delete(ctx context.Context, bucket, key string) error {
	_, err := b.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	return err
}


func (b *BlobStore) CreateAndSaveNewsletterGift(ctx context.Context, name string) (string, error) {
	// We write the image to an intermediate buffer. The image is a few hundred kB, so that's fine.
	var buffer bytes.Buffer

	w := Wallpaper{Name: name}
	
	if err := w.Generate(&buffer, time.Now().Unix()); err != nil {
		return "", fmt.Errorf("error generating wallpaper image: %w", err)
	}

	// Just use a UUIDv4 for the key, so we avoid collisions and don't have to sanitize the name further
	key := fmt.Sprintf("gifts/%v.png", uuid.NewString())

	if err := b.Put(ctx, b.bucket, key, "image/png", bytes.NewReader(buffer.Bytes())); err != nil {
		return "", fmt.Errorf("error putting wallpaper image: %w", err)
	}

	// Create a presigned URL that allows access to the image for 7 days
	presignClient := s3.NewPresignClient(b.Client, func(o *s3.PresignOptions) {
		o.Expires = 7 * 24 * time.Hour
	})
	presignedRequest, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &b.bucket,
		Key:    &key,
	})
	if err != nil {
		return "", fmt.Errorf("error creating presigned url for wallpaper image: %w", err)
	}
	return presignedRequest.URL, nil
}
