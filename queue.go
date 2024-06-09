package queue

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Queue struct {
	db DbInterface
}

const (
	StatePending   = "pending"
	StateRunning   = "running"
	StateCompleted = "completed"
	StateError     = "error"
)

const (
	DefaultTimeout  = time.Minute * 5
	DefaultMaxTries = 3
)

type Meta struct {
	Created    time.Time  `bson:"created"`
	Dispatched *time.Time `bson:"dispatched"`
	Completed  *time.Time `bson:"completed"`
}

type Task struct {
	Id       primitive.ObjectID `bson:"_id,omitempty"`
	Topic    string             `bson:"topic"`
	Payload  any                `bson:"payload"`
	Tries    uint               `bson:"tries"`
	MaxTries uint               `bson:"maxtries"`
	State    string             `bson:"state"`
	Message  string             `bson:"message"`
	Meta     Meta
}

func NewQueue(db DbInterface) *Queue {
	queue := Queue{
		db: db,
	}

	return &queue
}

var nowFunc = time.Now

func setNowFunc(n func() time.Time) {
	nowFunc = n
}

func (q *Queue) Publish(topic string, payload any, maxTries uint) (*Task, error) {
	if maxTries == 0 {
		maxTries = DefaultMaxTries
	}

	t := Task{
		Topic:    topic,
		Payload:  payload,
		Tries:    0,
		MaxTries: maxTries,
		Meta: Meta{
			Created:    nowFunc(),
			Dispatched: nil,
			Completed:  nil,
		},
		State: StatePending,
	}

	insertedId, err := q.db.InsertOne(t)
	if err != nil {
		return nil, err
	}

	t.Id = insertedId

	return &t, nil
}

func (q *Queue) GetNext(topic string) (*Task, error) {
	t := Task{}
	res := q.db.FindOneAndUpdate(bson.M{
		"topic": topic,
		"state": StatePending,
		"$expr": bson.M{"$lt": bson.A{"$tries", "$maxtries"}},
	},
		bson.M{
			"$set": bson.M{"state": StateRunning, "meta.dispatched": time.Now()},
			"$inc": bson.M{"tries": 1},
		})

	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return nil, nil
	}

	if err := res.Decode(&t); err != nil {
		return nil, err
	}

	return &t, nil
}

type Callback func(t Task)

func (q *Queue) Subscribe(topic string, cb Callback) error {
	pipeline := bson.D{
		{"$match", bson.D{{"operationType", "insert"}, {"fullDocument.topic", topic}, {"state", StatePending}}},
	}

	stream, err := q.db.Watch(mongo.Pipeline{pipeline})
	if err != nil {
		return err
	}

	defer stream.Close(q.db.Context())

	for stream.Next(q.db.Context()) {
		var event struct {
			Task Task `bson:"fullDocument"`
		}

		if err := stream.Decode(&event); err != nil {
			continue
		}

		event.Task.State = StateRunning
		now := time.Now()
		event.Task.Meta.Dispatched = &now

		err := q.db.UpdateOne(
			bson.M{"_id": event.Task.Id},
			bson.M{"$set": bson.M{
				"state":           event.Task.State,
				"meta.dispatched": event.Task.Meta.Dispatched,
			}})

		if err != nil {
			_ = q.Err(event.Task.Id.Hex(), err)
			continue
		}

		cb(event.Task)
	}

	return nil
}

func (q *Queue) Ack(id string) error {
	oId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	return q.db.UpdateOne(
		bson.M{"_id": oId},
		bson.M{"$set": bson.M{
			"state":          StateCompleted,
			"meta.completed": time.Now(),
		}})
}

func (q *Queue) Err(id string, err error) error {
	oId, e := primitive.ObjectIDFromHex(id)
	if e != nil {
		return err
	}

	return q.db.UpdateOne(
		bson.M{"_id": oId},
		bson.M{"$set": bson.M{
			"state":          StateError,
			"meta.completed": time.Now(),
			"message":        err.Error()},
		})
}

func (q *Queue) Selfcare(topic *string) error {
	// re-schedule long-running tasks
	// this only happens if the processor could not ack the task, i.e. the application crashed
	query := bson.M{
		"state":           StateRunning,
		"meta.dispatched": bson.M{"$lt": time.Now().Add(DefaultTimeout)},
	}
	if topic != nil {
		query["topic"] = *topic
	}

	_ = q.db.UpdateMany(
		query,
		bson.M{"$set": bson.M{
			"state":           StatePending,
			"meta.dispatched": nil},
		})

	// set tasks exceeding maxtries to error
	query = bson.M{
		"state": StatePending,
		"$expr": bson.M{"$gte": bson.A{"$tries", "$maxtries"}},
	}
	if topic != nil {
		query["topic"] = *topic
	}

	_ = q.db.UpdateMany(
		query,
		bson.M{"$set": bson.M{
			"state":          StateError,
			"meta.completed": time.Now()},
		})

	return nil
}

func (q *Queue) CreateIndexes() error {
	err := q.db.CreateIndexes([]mongo.IndexModel{{
		Keys: bson.D{{"topic", 1}, {"state", 1}},
	}, {
		Keys: bson.D{{"meta.completed", 1}}, Options: options.Index().SetExpireAfterSeconds(3600),
	}})

	return err
}
