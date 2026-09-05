package queue

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ChangeStreamInterface interface {
	Next(ctx context.Context) bool
	Decode(v any) error
	Close(ctx context.Context) error
}

type CollectionInterface interface {
	InsertOne(ctx context.Context, document any, opts ...options.Lister[options.InsertOneOptions]) (res *mongo.InsertOneResult, err error)
	FindOneAndUpdate(ctx context.Context, filter any, update any, opts ...options.Lister[options.FindOneAndUpdateOptions]) *mongo.SingleResult
	UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (res *mongo.UpdateResult, err error)
	UpdateMany(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (res *mongo.UpdateResult, err error)
	Watch(ctx context.Context, pipeline any, opts ...options.Lister[options.ChangeStreamOptions]) (stream *mongo.ChangeStream, err error)
	Indexes() mongo.IndexView
}

type StdDb struct {
	collection CollectionInterface
}

func NewStdDb(collection CollectionInterface) *StdDb {

	db := StdDb{
		collection: collection,
	}

	return &db
}

func (d *StdDb) InsertOne(ctx context.Context, document any) (bson.ObjectID, error) {
	res, err := d.collection.InsertOne(ctx, document)
	if err != nil {
		return bson.NilObjectID, err
	}

	return res.InsertedID.(bson.ObjectID), nil
}

func (d *StdDb) FindOneAndUpdate(ctx context.Context, filter any, update any, opts ...options.Lister[options.FindOneAndUpdateOptions]) *mongo.SingleResult {
	opts = append(opts, options.FindOneAndUpdate().SetReturnDocument(options.After))

	res := d.collection.FindOneAndUpdate(ctx, filter, update, opts...)
	if res == nil {
		return mongo.NewSingleResultFromDocument(bson.M{}, errors.New("no result returned"), nil)
	}
	return res
}

func (d *StdDb) UpdateOne(ctx context.Context, filter any, update any) error {
	_, err := d.collection.UpdateOne(ctx, filter, update)
	return err
}

func (d *StdDb) UpdateMany(ctx context.Context, filter any, update any) error {
	_, err := d.collection.UpdateMany(ctx, filter, update)
	return err
}

func (d *StdDb) Watch(ctx context.Context, pipeline any) (ChangeStreamInterface, error) {
	return d.collection.Watch(ctx, pipeline)
}

func (d *StdDb) CreateIndexes(ctx context.Context, indexes []mongo.IndexModel) error {
	_, err := d.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
