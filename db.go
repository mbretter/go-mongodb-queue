package queue

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbInterface interface {
	InsertOne(document interface{}) (primitive.ObjectID, error)
	FindOneAndUpdate(filter interface{}, update interface{}) *mongo.SingleResult
	UpdateOne(filter interface{}, update interface{}) error
	UpdateMany(filter interface{}, update interface{}) error
	Watch(pipeline interface{}) (*mongo.ChangeStream, error)
	CreateIndexes(index []mongo.IndexModel) error
	Context() context.Context
}

type StdDb struct {
	context    context.Context
	collection *mongo.Collection
}

func NewStdDb(collection *mongo.Collection, ctx context.Context) *StdDb {
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

func (d *StdDb) FindOneAndUpdate(filter interface{}, update interface{}) *mongo.SingleResult {
	res := d.collection.FindOneAndUpdate(d.context, filter, update,
		options.FindOneAndUpdate().SetReturnDocument(options.After))
	if res == nil {
		return mongo.NewSingleResultFromDocument(nil, errors.New("no result returned"), nil)
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

func (d *StdDb) Watch(pipeline interface{}) (*mongo.ChangeStream, error) {
	return d.collection.Watch(d.context, pipeline)
}

func (d *StdDb) CreateIndexes(indexes []mongo.IndexModel) error {
	_, err := d.collection.Indexes().CreateMany(d.context, indexes)
	return err
}

func (d *StdDb) Context() context.Context {
	return d.context
}
