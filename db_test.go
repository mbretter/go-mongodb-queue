package queue

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestDb_NewStd(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "Success without context",
		},
		{
			name: "Success with context",
			ctx:  context.TODO(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collectionMock := NewCollectionInterfaceMock(t)
			db := NewStdDb(collectionMock)

			if assert.NotNil(t, db) {
				assert.Equal(t, db.collection, collectionMock)
			}
		})
	}
}

func TestDb_InsertOne(t *testing.T) {
	tests := []struct {
		name  string
		error error
	}{
		{
			name: "Success",
		},
		{
			name:  "Error",
			error: errors.New("insert failed"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collectionMock := NewCollectionInterfaceMock(t)
			db := NewStdDb(collectionMock)

			doc := bson.M{"foo": "bar"}
			res := mongo.InsertOneResult{
				InsertedID: bson.NewObjectID(),
			}
			collectionMock.EXPECT().InsertOne(context.TODO(), doc).Return(&res, tt.error)

			oId, err := db.InsertOne(context.TODO(), doc)

			assert.Equal(t, err, tt.error)
			if tt.error == nil {
				assert.Equal(t, oId, res.InsertedID)
			} else {
				assert.Equal(t, oId, bson.NilObjectID)
			}

		})
	}
}

func TestDb_FindOneAndUpdate(t *testing.T) {
	tests := []struct {
		name string
		res  *mongo.SingleResult
	}{
		{
			name: "Success",
			res:  &mongo.SingleResult{},
		},
		{
			name: "No doc found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collectionMock := NewCollectionInterfaceMock(t)
			db := NewStdDb(collectionMock)

			filter := bson.M{"foo": "bar"}
			upd := bson.M{"status": "active"}

			collectionMock.EXPECT().FindOneAndUpdate(context.TODO(), filter, upd, mock.MatchedBy(func(opts *options.FindOneAndUpdateOptionsBuilder) bool {
				if opts == nil {
					return false
				}

				resolvedOpts := &options.FindOneAndUpdateOptions{}
				for _, f := range opts.Opts {
					if err := f(resolvedOpts); err != nil {
						return false
					}
				}
				return resolvedOpts.ReturnDocument != nil && *resolvedOpts.ReturnDocument == options.After
			})).Return(tt.res)

			res := db.FindOneAndUpdate(context.TODO(), filter, upd)

			if tt.res == nil {
				assert.Equal(t, errors.New("no result returned"), res.Err())
			} else {
				assert.Equal(t, tt.res, res)
			}
		})
	}
}

func TestDb_UpdateOne(t *testing.T) {
	tests := []struct {
		name  string
		error error
	}{
		{
			name: "Success",
		},
		{
			name:  "Error",
			error: errors.New("update failed"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collectionMock := NewCollectionInterfaceMock(t)
			db := NewStdDb(collectionMock)

			filter := bson.M{"foo": "bar"}
			upd := bson.M{"status": "active"}

			collectionMock.EXPECT().UpdateOne(context.TODO(), filter, upd).Return(&mongo.UpdateResult{}, tt.error)

			err := db.UpdateOne(context.TODO(), filter, upd)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestDb_UpdateMany(t *testing.T) {
	tests := []struct {
		name  string
		error error
	}{
		{
			name: "Success",
		},
		{
			name:  "Error",
			error: errors.New("update failed"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collectionMock := NewCollectionInterfaceMock(t)
			db := NewStdDb(collectionMock)

			filter := bson.M{"foo": "bar"}
			upd := bson.M{"status": "active"}

			collectionMock.EXPECT().UpdateMany(context.TODO(), filter, upd).Return(&mongo.UpdateResult{}, tt.error)

			err := db.UpdateMany(context.TODO(), filter, upd)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestDb_Watch(t *testing.T) {
	tests := []struct {
		name  string
		error error
	}{
		{
			name: "Success",
		},
		{
			name:  "Error",
			error: errors.New("watch failed"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collectionMock := NewCollectionInterfaceMock(t)
			db := NewStdDb(collectionMock)

			pipeline := mongo.Pipeline{bson.D{{Key: "$match", Value: bson.M{"foo": "bar"}}}}

			changeStream := mongo.ChangeStream{}

			collectionMock.EXPECT().Watch(context.TODO(), pipeline).Return(&changeStream, tt.error)

			cs, err := db.Watch(context.TODO(), pipeline)

			assert.Equal(t, tt.error, err)
			assert.Implements(t, new(ChangeStreamInterface), cs)
		})
	}
}
