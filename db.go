package queue

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DbInterface interface {
	InsertOne(document any) (bson.ObjectID, error)
	FindOneAndUpdate(filter any, update any, opts ...options.Lister[options.FindOneAndUpdateOptions]) *mongo.SingleResult
	UpdateOne(filter any, update any) error
	UpdateMany(filter any, update any) error
	Watch(pipeline any) (ChangeStreamInterface, error)
	CreateIndexes(index []mongo.IndexModel) error
	Context() context.Context
}

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
	context    context.Context
	collection CollectionInterface
}

func NewStdDb(collection CollectionInterface, ctx context.Context) *StdDb {
	if ctx == nil {
		ctx = context.Background()
	}

	db := StdDb{
		context:    ctx,
		collection: collection,
	}

	return &db
}

func (d *StdDb) InsertOne(document any) (bson.ObjectID, error) {
	res, err := d.collection.InsertOne(d.context, document)
	if err != nil {
		return bson.NilObjectID, err
	}

	return res.InsertedID.(bson.ObjectID), nil
}

func (d *StdDb) FindOneAndUpdate(filter any, update any, opts ...options.Lister[options.FindOneAndUpdateOptions]) *mongo.SingleResult {
	opts = append(opts, options.FindOneAndUpdate().SetReturnDocument(options.After))

	res := d.collection.FindOneAndUpdate(d.context, filter, update, opts...)
	if res == nil {
		return mongo.NewSingleResultFromDocument(bson.M{}, errors.New("no result returned"), nil)
	}
	return res
}

func (d *StdDb) UpdateOne(filter any, update any) error {
	_, err := d.collection.UpdateOne(d.context, filter, update)
	return err
}

func (d *StdDb) UpdateMany(filter any, update any) error {
	_, err := d.collection.UpdateMany(d.context, filter, update)
	return err
}

func (d *StdDb) Watch(pipeline any) (ChangeStreamInterface, error) {
	return d.collection.Watch(d.context, pipeline)
}

func (d *StdDb) CreateIndexes(indexes []mongo.IndexModel) error {
	_, err := d.collection.Indexes().CreateMany(d.context, indexes)
	return err
}

func (d *StdDb) Context() context.Context {
	return d.context
}
