package repository

import (
	"context"
	"io"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepo struct {
	Bucket *gridfs.Bucket
}

func NewMongoRepo(uri, dbName string) (*MongoRepo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	db := client.Database(dbName)

	bucket, err := gridfs.NewBucket(db)
	if err != nil {
		return nil, err
	}

	return &MongoRepo{Bucket: bucket}, nil
}

func (m *MongoRepo) UploadFile(ctx context.Context, path, name string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	stream, err := m.Bucket.OpenUploadStream(name)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	_, err = io.Copy(stream, file)
	if err != nil {
		return "", err
	}

	return stream.FileID.(primitive.ObjectID).Hex(), nil
}

func (m *MongoRepo) DownloadFile(ctx context.Context, name string, w io.Writer) error {
	_, err := m.Bucket.DownloadToStreamByName(name, w)
	return err
}
