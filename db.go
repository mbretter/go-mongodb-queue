package queue

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbInterface interface {
	InsertOne(document interface{}) (primitive.ObjectID, error)
	FindOneAndUpdate(filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult
	UpdateOne(filter interface{}, update interface{}) error
	UpdateMany(filter interface{}, update interface{}) error
	Watch(pipeline interface{}) (ChangeStreamInterface, error)
	CreateIndexes(index []mongo.IndexModel) error
	Context() context.Context
}

type ChangeStreamInterface interface {
	Next(ctx context.Context) bool
	Decode(v interface{}) error
	Close(ctx context.Context) error
}

type CollectionInterface interface {
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error)
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

func (d *StdDb) InsertOne(document interface{}) (primitive.ObjectID, error) {
	res, err := d.collection.InsertOne(d.context, document)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return res.InsertedID.(primitive.ObjectID), nil
}

func (d *StdDb) FindOneAndUpdate(filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {
	opts = append(opts, options.FindOneAndUpdate().SetReturnDocument(options.After))
	res := d.collection.FindOneAndUpdate(d.context, filter, update, opts...)
	if res == nil {
		return mongo.NewSingleResultFromDocument(bson.M{}, errors.New("no result returned"), nil)
	}
	return res
}

func (d *StdDb) UpdateOne(filter interface{}, update interface{}) error {
	_, err := d.collection.UpdateOne(d.context, filter, update)
	return err
}

func (d *StdDb) UpdateMany(filter interface{}, update interface{}) error {
	_, err := d.collection.UpdateMany(d.context, filter, update)
	return err
}

func (d *StdDb) Watch(pipeline interface{}) (ChangeStreamInterface, error) {
	return d.collection.Watch(d.context, pipeline)
}

func (d *StdDb) CreateIndexes(indexes []mongo.IndexModel) error {
	_, err := d.collection.Indexes().CreateMany(d.context, indexes)
	return err
}

func (d *StdDb) Context() context.Context {
	return d.context
}
