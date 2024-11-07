package queue

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
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

			oId := primitive.NewObjectID()

			taskExpected := Task{
				Topic:    tt.topic,
				Payload:  tt.payload,
				Tries:    0,
				MaxTries: 3,
				Meta: Meta{
					Created: nowFunc(),
				},
				State: StatePending,
			}
			dbMock.EXPECT().InsertOne(taskExpected).Return(oId, tt.error)

			task, err := q.Publish(tt.topic, tt.payload, tt.maxTries)

			if tt.error == nil {
				taskExpected.Id = oId
				assert.Equal(t, taskExpected, *task)
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
		task        *Task
		watchError  error
		decodeError error
		updateError error
	}{
		{
			name:  "Success",
			topic: "topic1",
			task: &Task{
				Id:       primitive.NewObjectID(),
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
			task: &Task{
				Id:       primitive.NewObjectID(),
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
			task: &Task{
				Id:       primitive.NewObjectID(),
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
			task: &Task{
				Id:       primitive.NewObjectID(),
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
				{"$match", bson.D{{"operationType", "insert"}, {"fullDocument.topic", tt.topic}, {"fullDocument.state", StatePending}}},
			}
			changeStream := NewChangeStreamInterfaceMock(t)
			dbMock.EXPECT().Watch(mongo.Pipeline{pipeline}).Return(changeStream, tt.watchError)

			res := mongo.NewSingleResultFromDocument(Task{}, mongo.ErrNoDocuments, nil)

			if tt.watchError != nil {
				goto runTest
			}

			changeStream.EXPECT().Close(context.TODO()).Return(nil)
			dbMock.EXPECT().Context().Return(context.TODO())

			dbMock.EXPECT().FindOneAndUpdate(bson.M{
				"topic": tt.topic,
				"state": StatePending,
				"$expr": bson.M{"$lt": bson.A{"$tries", "$maxtries"}},
			}, bson.M{
				"$set": bson.M{"state": StateRunning, "meta.dispatched": now},
				"$inc": bson.M{"tries": 1},
			}, options.FindOneAndUpdate().SetSort(bson.D{{"meta.scheduled", 1}})).Return(res)

			if tt.task != nil {
				changeStream.EXPECT().Next(context.TODO()).Once().Return(true)
				changeStream.EXPECT().Next(context.TODO()).Return(false)
				var evt event
				changeStream.EXPECT().Decode(&evt).RunAndReturn(func(i interface{}) error {
					i.(*event).Task = *tt.task
					return tt.decodeError
				})

				if tt.decodeError != nil {
					goto runTest
				}

				if tt.name == "AlreadyProcessed" {
					goto runTest
				}

				dbMock.EXPECT().UpdateOne(bson.M{"_id": tt.task.Id},
					bson.M{"$set": bson.M{
						"state":           StateRunning,
						"meta.dispatched": &now,
					}}).Return(tt.updateError)

				if tt.updateError != nil {
					dbMock.EXPECT().UpdateOne(
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
			err := q.Subscribe(tt.topic, func(task Task) {
				assert.Equal(t, StateRunning, task.State)
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
		task  Task
		error error
	}{
		{
			name:  "Success",
			topic: "topic1",
			task: Task{
				Id:       primitive.NewObjectID(),
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
			task:  Task{},
			error: errors.New("FindOneAndUpdate failed"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbMock := NewDbInterfaceMock(t)
			q := NewQueue(dbMock)

			pipeline := bson.D{
				{"$match", bson.D{{"operationType", "insert"}, {"fullDocument.topic", tt.topic}, {"fullDocument.state", StatePending}}},
			}
			changeStream := NewChangeStreamInterfaceMock(t)
			dbMock.EXPECT().Watch(mongo.Pipeline{pipeline}).Return(changeStream, nil)

			changeStream.EXPECT().Close(context.TODO()).Return(nil)
			dbMock.EXPECT().Context().Return(context.TODO())

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

			opts := options.FindOneAndUpdate().SetSort(bson.D{{"meta.scheduled", 1}})
			dbMock.EXPECT().FindOneAndUpdate(filter, update, opts).Once().Return(res)

			if tt.error == nil {
				dbMock.EXPECT().FindOneAndUpdate(filter, update, opts).Once().Return(resNoDoc)
				changeStream.EXPECT().Next(context.TODO()).Return(false)
			}

			err := q.Subscribe(tt.topic, func(task Task) {
				assert.Equal(t, StateRunning, task.State)
			})

			assert.Equal(t, tt.error, err)
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

			oId, err := primitive.ObjectIDFromHex(tt.taskId)

			if err == nil {
				dbMock.EXPECT().UpdateOne(
					bson.M{"_id": oId},
					bson.M{"$set": bson.M{
						"state":          StateCompleted,
						"meta.completed": nowFunc(),
					}}).Return(tt.error)
			}

			err = q.Ack(tt.taskId)

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

			oId, err := primitive.ObjectIDFromHex(tt.taskId)

			if err == nil {
				dbMock.EXPECT().UpdateOne(
					bson.M{"_id": oId},
					bson.M{"$set": bson.M{
						"state":          StateError,
						"meta.completed": nowFunc(),
						"message":        "some error",
					}}).Return(tt.error)
			}

			err = q.Err(tt.taskId, errors.New("some error"))

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

			dbMock.EXPECT().UpdateMany(query1,
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

			dbMock.EXPECT().UpdateMany(query2,
				bson.M{"$set": bson.M{
					"state":          StateError,
					"meta.completed": nowFunc()},
				}).Return(tt.error2)

			err := q.Selfcare(tt.topic, 0)

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

			dbMock.EXPECT().CreateIndexes([]mongo.IndexModel{{
				Keys: bson.D{{"topic", 1}, {"state", 1}},
			}, {
				Keys: bson.D{{"meta.completed", 1}}, Options: options.Index().SetExpireAfterSeconds(3600),
			}}).Return(tt.error)

			err := q.CreateIndexes()
			assert.Equal(t, err, tt.error)
		})
	}
}
