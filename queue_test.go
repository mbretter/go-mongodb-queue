package queue

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestQueue_Publish(t *testing.T) {
	setNowFunc(func() time.Time {
		t, _ := time.Parse(time.DateTime, "2023-11-12 15:04:05")
		return t
	})

	tests := []struct {
		name     string
		topic    string
		payload  any
		maxTries uint
		error    error
	}{
		{
			name:     "Success",
			topic:    "topic1",
			payload:  "payload1",
			maxTries: 3,
		},
		{
			name:     "Error",
			topic:    "topic2",
			payload:  "payload2",
			maxTries: 0,
			error:    errors.New("db insert failed"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbMock := NewDbInterfaceMock(t)
			q := NewQueue(dbMock)

			oId := bson.NewObjectID()

			taskExpected := task{
				Topic:    tt.topic,
				Payload:  tt.payload,
				Tries:    0,
				MaxTries: 3,
				Meta: Meta{
					Created: nowFunc(),
				},
				State: StatePending,
			}
			dbMock.EXPECT().InsertOne(context.TODO(), taskExpected).Return(oId, tt.error)

			opts := NewPublishOptions().SetMaxTries(tt.maxTries)
			task, err := q.Publish(context.TODO(), tt.topic, tt.payload, opts, nil)

			if tt.error == nil {
				taskExpected.Id = oId
				assert.Equal(t, &taskExpected, task)
			} else {
				assert.Nil(t, task)
				assert.Equal(t, tt.error, err)
			}

		})
	}
}

func TestQueue_Subscribe(t *testing.T) {
	setNowFunc(func() time.Time {
		t, _ := time.Parse(time.DateTime, "2024-10-12 15:04:05")
		return t
	})

	now := nowFunc()

	tests := []struct {
		name        string
		topic       string
		task        *task
		watchError  error
		decodeError error
		updateError error
	}{
		{
			name:  "Success",
			topic: "topic1",
			task: &task{
				Id:       bson.NewObjectID(),
				Topic:    "topic1",
				Payload:  "payload1",
				Tries:    1,
				MaxTries: 3,
				Meta: Meta{
					Created: now,
				},
				State: StatePending,
			},
		},
		{
			name:       "WatchError",
			topic:      "topic1",
			watchError: errors.New("watch failed"),
		},
		{
			name:  "EventDecodeError",
			topic: "topic1",
			task: &task{
				Id:       bson.NewObjectID(),
				Topic:    "topic1",
				Payload:  "payload1",
				Tries:    1,
				MaxTries: 3,
				Meta: Meta{
					Created: now,
				},
				State: StatePending,
			},
			decodeError: errors.New("decode failed"),
		},
		{
			name:  "AlreadyProcessed",
			topic: "topic1",
			task: &task{
				Id:       bson.NewObjectID(),
				Topic:    "topic1",
				Payload:  "payload1",
				Tries:    1,
				MaxTries: 3,
				Meta: Meta{
					Created: now.Add(-time.Hour),
				},
				State: StatePending,
			},
		},
		{
			name:  "UpdateError",
			topic: "topic1",
			task: &task{
				Id:       bson.NewObjectID(),
				Topic:    "topic1",
				Payload:  "payload1",
				Tries:    1,
				MaxTries: 3,
				Meta: Meta{
					Created: nowFunc(),
				},
				State: StatePending,
			},
			updateError: errors.New("update failed"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbMock := NewDbInterfaceMock(t)
			q := NewQueue(dbMock)

			pipeline := bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "operationType", Value: "insert"},
					{Key: "fullDocument.topic", Value: tt.topic},
					{Key: "fullDocument.state", Value: StatePending},
				}},
			}
			changeStream := NewChangeStreamInterfaceMock(t)
			dbMock.EXPECT().Watch(context.TODO(), mongo.Pipeline{pipeline}).Return(changeStream, tt.watchError)

			res := mongo.NewSingleResultFromDocument(&task{}, mongo.ErrNoDocuments, nil)

			if tt.watchError != nil {
				goto runTest
			}

			changeStream.EXPECT().Close(context.TODO()).Return(nil)
			//options.FindOneAndUpdate().SetSort(bson.M{"meta.scheduled": 1}).SetReturnDocument(options.After))
			dbMock.EXPECT().FindOneAndUpdate(context.TODO(), bson.M{
				"topic": tt.topic,
				"state": StatePending,
				"$expr": bson.M{"$lt": bson.A{"$tries", "$maxtries"}},
			}, bson.M{
				"$set": bson.M{"state": StateRunning, "meta.dispatched": now},
				"$inc": bson.M{"tries": 1},
			}, mock.MatchedBy(func(opts *options.FindOneAndUpdateOptionsBuilder) bool {
				if opts == nil {
					return false
				}

				resolvedOpts := &options.FindOneAndUpdateOptions{}
				for _, f := range opts.Opts {
					if err := f(resolvedOpts); err != nil {
						return false
					}
				}
				return resolvedOpts.ReturnDocument != nil && *resolvedOpts.ReturnDocument == options.After &&
					resolvedOpts.Sort != nil && reflect.DeepEqual(resolvedOpts.Sort, bson.M{"meta.scheduled": 1})
			})).Once().Return(res)

			if tt.task != nil {
				changeStream.EXPECT().Next(context.TODO()).Once().Return(true)
				changeStream.EXPECT().Next(context.TODO()).Return(false)
				var evt event
				changeStream.EXPECT().Decode(&evt).RunAndReturn(func(i any) error {
					i.(*event).Task = tt.task
					return tt.decodeError
				})

				if tt.decodeError != nil {
					goto runTest
				}

				if tt.name == "AlreadyProcessed" {
					goto runTest
				}

				retTask := *tt.task
				retTask.State = StateRunning
				res = mongo.NewSingleResultFromDocument(retTask, tt.updateError, nil)

				dbMock.EXPECT().FindOneAndUpdate(context.TODO(), bson.M{
					"_id":   tt.task.Id,
					"state": StatePending,
					"$expr": bson.M{"$lt": bson.A{"$tries", "$maxtries"}},
				}, bson.M{
					"$set": bson.M{"state": StateRunning, "meta.dispatched": now},
					"$inc": bson.M{"tries": 1},
				}, stdOptsMatcher).Once().Return(res)

				if tt.updateError != nil {
					dbMock.EXPECT().UpdateOne(context.TODO(),
						bson.M{"_id": tt.task.Id},
						bson.M{"$set": bson.M{
							"state":          StateError,
							"meta.completed": nowFunc(),
							"message":        tt.updateError.Error()},
						}).Return(nil)
				}
			} else {
				changeStream.EXPECT().Next(context.TODO()).Return(false)
			}

		runTest:
			err := q.Subscribe(context.TODO(), tt.topic, func(task Task) {
				assert.Equal(t, StateRunning, task.GetState())
			})

			if tt.watchError != nil {
				assert.Equal(t, tt.watchError, err)
			} else {
				assert.Nil(t, err)
			}

		})
	}
}

func TestQueue_SubscribeUnprocessedTasks(t *testing.T) {
	setNowFunc(func() time.Time {
		t, _ := time.Parse(time.DateTime, "2024-10-12 15:04:05")
		return t
	})

	now := nowFunc()

	tests := []struct {
		name  string
		topic string
		task  *task
		error error
	}{
		{
			name:  "Success",
			topic: "topic1",
			task: &task{
				Id:       bson.NewObjectID(),
				Topic:    "topic1",
				Payload:  "payload1",
				Tries:    1,
				MaxTries: 3,
				Meta: Meta{
					Created: now,
				},
				State: StateRunning,
			},
		},
		{
			name:  "Error",
			topic: "topic1",
			task:  &task{},
			error: errors.New("FindOneAndUpdate failed"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbMock := NewDbInterfaceMock(t)
			q := NewQueue(dbMock)

			pipeline := bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "operationType", Value: "insert"},
					{Key: "fullDocument.topic", Value: tt.topic},
					{Key: "fullDocument.state", Value: StatePending},
				}},
			}
			changeStream := NewChangeStreamInterfaceMock(t)
			dbMock.EXPECT().Watch(context.TODO(), mongo.Pipeline{pipeline}).Return(changeStream, nil)

			changeStream.EXPECT().Close(context.TODO()).Return(nil)

			res := mongo.NewSingleResultFromDocument(tt.task, tt.error, nil)
			resNoDoc := mongo.NewSingleResultFromDocument(tt.task, mongo.ErrNoDocuments, nil)

			filter := bson.M{
				"topic": tt.topic,
				"state": StatePending,
				"$expr": bson.M{"$lt": bson.A{"$tries", "$maxtries"}},
			}

			update := bson.M{
				"$set": bson.M{"state": StateRunning, "meta.dispatched": now},
				"$inc": bson.M{"tries": 1},
			}

			optsMatcher := mock.MatchedBy(func(opts *options.FindOneAndUpdateOptionsBuilder) bool {
				if opts == nil {
					return false
				}

				resolvedOpts := &options.FindOneAndUpdateOptions{}
				for _, f := range opts.Opts {
					if err := f(resolvedOpts); err != nil {
						return false
					}
				}
				return resolvedOpts.ReturnDocument != nil && *resolvedOpts.ReturnDocument == options.After &&
					resolvedOpts.Sort != nil && reflect.DeepEqual(resolvedOpts.Sort, bson.M{"meta.scheduled": 1})
			})

			dbMock.EXPECT().FindOneAndUpdate(context.TODO(), filter, update, optsMatcher).Once().Return(res)

			if tt.error == nil {
				dbMock.EXPECT().FindOneAndUpdate(context.TODO(), filter, update, optsMatcher).Once().Return(resNoDoc)
				changeStream.EXPECT().Next(context.TODO()).Return(false)
			}

			err := q.Subscribe(context.TODO(), tt.topic, func(task Task) {
				assert.Equal(t, StateRunning, task.GetState())
			})

			assert.Equal(t, tt.error, err)
		})
	}
}

func TestQueue_GetNextById(t *testing.T) {
	setNowFunc(func() time.Time {
		t, _ := time.Parse(time.DateTime, "2024-10-12 15:04:05")
		return t
	})

	now := nowFunc()

	tests := []struct {
		name  string
		task  *task
		error error
	}{
		{
			name: "Success",
			task: &task{
				Id:       bson.NewObjectID(),
				Topic:    "topic1",
				Payload:  "payload1",
				Tries:    1,
				MaxTries: 3,
				Meta: Meta{
					Created: now,
				},
				State: StateRunning,
			},
		},
		{
			name: "Error",
			task: &task{
				Id: bson.NewObjectID(),
			},
			error: errors.New("no doc found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbMock := NewDbInterfaceMock(t)
			q := NewQueue(dbMock)

			resOk := mongo.NewSingleResultFromDocument(tt.task, tt.error, nil)
			resNoDoc := mongo.NewSingleResultFromDocument(tt.task, mongo.ErrNoDocuments, nil)

			var res *mongo.SingleResult
			if tt.error == nil {
				res = resOk
			} else {
				res = resNoDoc
			}

			filter := bson.M{
				"_id":   tt.task.Id,
				"state": StatePending,
				"$expr": bson.M{"$lt": bson.A{"$tries", "$maxtries"}},
			}

			update := bson.M{
				"$set": bson.M{"state": StateRunning, "meta.dispatched": now},
				"$inc": bson.M{"tries": 1},
			}

			dbMock.EXPECT().FindOneAndUpdate(context.TODO(), filter, update, stdOptsMatcher).Once().Return(res)

			ts, err := q.GetNextById(context.TODO(), tt.task.Id)

			if tt.error == nil {
				assert.Equal(t, tt.task.Topic, ts.GetTopic())
				assert.Equal(t, tt.error, err)
			} else {
				assert.Nil(t, ts)
				assert.Nil(t, err)
			}

		})
	}
}

func TestQueue_Ack(t *testing.T) {
	setNowFunc(func() time.Time {
		t, _ := time.Parse(time.DateTime, "2024-10-12 15:04:05")
		return t
	})

	tests := []struct {
		name   string
		taskId string
		error  error
	}{
		{
			name:   "Success",
			taskId: "67211cb175b7564a5cd9ce3f",
		},
		{
			name:   "InvalidObjectId",
			taskId: "xxx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbMock := NewDbInterfaceMock(t)

			q := NewQueue(dbMock)

			oId, err := bson.ObjectIDFromHex(tt.taskId)

			if err == nil {
				dbMock.EXPECT().UpdateOne(context.TODO(),
					bson.M{"_id": oId},
					bson.M{"$set": bson.M{
						"state":          StateCompleted,
						"meta.completed": nowFunc(),
					}}).Return(tt.error)
			}

			err = q.Ack(context.TODO(), tt.taskId)

			if tt.name == "InvalidObjectId" {
				assert.Equal(t, "the provided hex string is not a valid ObjectID", err.Error())
			} else {
				assert.Equal(t, tt.error, err)
			}
		})
	}
}

func TestQueue_Err(t *testing.T) {
	setNowFunc(func() time.Time {
		t, _ := time.Parse(time.DateTime, "2024-10-12 15:04:05")
		return t
	})

	tests := []struct {
		name   string
		taskId string
		error  error
	}{
		{
			name:   "Success",
			taskId: "67211cb175b7564a5cd9ce3f",
		},
		{
			name:   "InvalidObjectId",
			taskId: "xxx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbMock := NewDbInterfaceMock(t)

			q := NewQueue(dbMock)

			oId, err := bson.ObjectIDFromHex(tt.taskId)

			if err == nil {
				dbMock.EXPECT().UpdateOne(context.TODO(),
					bson.M{"_id": oId},
					bson.M{"$set": bson.M{
						"state":          StateError,
						"meta.completed": nowFunc(),
						"message":        "some error",
					}}).Return(tt.error)
			}

			err = q.Err(context.TODO(), tt.taskId, errors.New("some error"))

			if tt.name == "InvalidObjectId" {
				assert.Equal(t, "the provided hex string is not a valid ObjectID", err.Error())
			} else {
				assert.Equal(t, tt.error, err)
			}
		})
	}
}

func TestQueue_Selftest(t *testing.T) {
	setNowFunc(func() time.Time {
		t, _ := time.Parse(time.DateTime, "2024-11-04 15:04:05")
		return t
	})

	tests := []struct {
		name   string
		topic  string
		error1 error
		error2 error
	}{
		{
			name:  "Success",
			topic: "",
		},
		{
			name:  "Success with topic",
			topic: "user.delete",
		},
		{
			name:   "Reschedule failed",
			topic:  "",
			error1: errors.New("FindOneAndUpdate1"),
		},
		{
			name:   "Set maxtries to error failed",
			topic:  "",
			error2: errors.New("FindOneAndUpdate2"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbMock := NewDbInterfaceMock(t)

			q := NewQueue(dbMock)

			query1 := bson.M{
				"state":           StateRunning,
				"meta.dispatched": bson.M{"$lt": nowFunc().Add(DefaultTimeout)},
			}

			if tt.topic != "" {
				query1["topic"] = tt.topic
			}

			dbMock.EXPECT().UpdateMany(context.TODO(), query1,
				bson.M{"$set": bson.M{
					"state":           StatePending,
					"meta.dispatched": nil},
				}).Return(tt.error1)

			query2 := bson.M{
				"state": StatePending,
				"$expr": bson.M{"$gte": bson.A{"$tries", "$maxtries"}},
			}

			if tt.topic != "" {
				query2["topic"] = tt.topic
			}

			dbMock.EXPECT().UpdateMany(context.TODO(), query2,
				bson.M{"$set": bson.M{
					"state":          StateError,
					"meta.completed": nowFunc()},
				}).Return(tt.error2)

			err := q.Selfcare(context.TODO(), tt.topic, 0)

			if tt.error1 != nil {
				assert.Equal(t, tt.error1, err)
			} else if tt.error2 != nil {
				assert.Equal(t, tt.error2, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestQueue_CreateIndexes(t *testing.T) {
	tests := []struct {
		name  string
		error error
	}{
		{
			name: "Success",
		},
		{
			name:  "Error",
			error: errors.New("CreateIndexes failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbMock := NewDbInterfaceMock(t)

			q := NewQueue(dbMock)

			dbMock.EXPECT().CreateIndexes(context.TODO(), mock.MatchedBy(func(indexes []mongo.IndexModel) bool {
				if len(indexes) != 2 {
					return false
				}

				idx0 := indexes[0]
				if !reflect.DeepEqual(idx0.Keys, bson.D{{Key: "topic", Value: 1}, {Key: "state", Value: 1}}) {
					return false
				}
				if idx0.Options != nil {
					return false
				}

				idx1 := indexes[1]
				if !reflect.DeepEqual(idx1.Keys, bson.D{{Key: "meta.completed", Value: 1}}) {
					return false
				}

				if idx1.Options == nil {
					return false
				}
				resolvedOpts := &options.IndexOptions{}
				for _, f := range idx1.Options.Opts {
					if err := f(resolvedOpts); err != nil {
						return false
					}
				}
				return resolvedOpts.ExpireAfterSeconds != nil && *resolvedOpts.ExpireAfterSeconds == 3600
			})).Return(tt.error)

			err := q.CreateIndexes(context.TODO())
			assert.Equal(t, err, tt.error)
		})
	}
}

func TestQueue_Reschedule(t *testing.T) {
	setNowFunc(func() time.Time {
		t, _ := time.Parse(time.DateTime, "2023-11-12 15:04:05")
		return t
	})

	tests := []struct {
		name  string
		task  *task
		error error
	}{
		{
			name: "Success",
			task: &task{
				Topic:    "foo.bar",
				Payload:  "whatever",
				Tries:    1,
				MaxTries: 3,
				Meta: Meta{
					Created: nowFunc(),
				},
				State: StatePending,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbMock := NewDbInterfaceMock(t)
			q := NewQueue(dbMock)

			oId := bson.NewObjectID()

			taskExpected := task{
				Topic:    tt.task.Topic,
				Payload:  tt.task.Payload,
				Tries:    1,
				MaxTries: 3,
				Meta: Meta{
					Created: nowFunc(),
				},
				State: StatePending,
			}
			dbMock.EXPECT().InsertOne(context.TODO(), taskExpected).Return(oId, tt.error)

			task, err := q.Reschedule(context.TODO(), tt.task)

			if tt.error == nil {
				taskExpected.Id = oId
				assert.Equal(t, &taskExpected, task)
			} else {
				assert.Nil(t, task)
				assert.Equal(t, tt.error, err)
			}

		})
	}
}

var stdOptsMatcher = mock.MatchedBy(func(opts *options.FindOneAndUpdateOptionsBuilder) bool {
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
})
