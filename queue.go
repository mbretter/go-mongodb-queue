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

type event struct {
	Task Task `bson:"fullDocument"`
}

// NewQueue initializes a new Queue instance with the provided DbInterface.
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

type PublishOptions struct {
	MaxTries uint
	Tries    int
}

func NewPublishOptions() *PublishOptions {
	return &PublishOptions{
		MaxTries: 0,
		Tries:    -1,
	}
}

func (p *PublishOptions) SetMaxTries(maxTries uint) *PublishOptions {
	p.MaxTries = maxTries
	return p
}

func (p *PublishOptions) setTries(tries uint) *PublishOptions {
	p.Tries = int(tries)
	return p
}

// Publish inserts a new task into the queue with the given topic, payload, and maxTries.
// If maxTries is zero, it defaults to DefaultMaxTries.
func (q *Queue) Publish(topic string, payload any, opts ...*PublishOptions) (*Task, error) {

	o := PublishOptions{
		MaxTries: DefaultMaxTries,
		Tries:    0,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if opt.MaxTries > 0 {
			o.MaxTries = opt.MaxTries
		}

		if opt.Tries >= 0 {
			o.Tries = opt.Tries
		}
	}

	t := Task{
		Topic:    topic,
		Payload:  payload,
		Tries:    uint(o.Tries),
		MaxTries: o.MaxTries,
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
			"$set": bson.M{"state": StateRunning, "meta.dispatched": nowFunc()},
			"$inc": bson.M{"tries": 1},
		},
		options.FindOneAndUpdate().SetSort(bson.D{{"meta.scheduled", 1}}).SetReturnDocument(options.After),
	)

	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return nil, nil
	}

	if err := res.Decode(&t); err != nil {
		return nil, err
	}

	return &t, nil
}

func (q *Queue) GetNextById(id primitive.ObjectID) (*Task, error) {
	t := Task{}
	res := q.db.FindOneAndUpdate(bson.M{
		"_id":   id,
		"state": StatePending,
		"$expr": bson.M{"$lt": bson.A{"$tries", "$maxtries"}},
	},
		bson.M{
			"$set": bson.M{"state": StateRunning, "meta.dispatched": nowFunc()},
			"$inc": bson.M{"tries": 1},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return nil, nil
	}

	if err := res.Decode(&t); err != nil {
		return nil, err
	}

	return &t, nil
}

func (q *Queue) Reschedule(task *Task) (*Task, error) {
	return q.Publish(task.Topic, task.Payload, NewPublishOptions().setTries(task.Tries).SetMaxTries(task.MaxTries))
}

type Callback func(t Task)

func (q *Queue) Subscribe(topic string, cb Callback) error {
	pipeline := bson.D{{"$match", bson.D{
		{"operationType", "insert"},
		{"fullDocument.topic", topic},
		{"fullDocument.state", StatePending}}},
	}

	stream, err := q.db.Watch(mongo.Pipeline{pipeline})
	if err != nil {
		return err
	}
	//goland:noinspection ALL
	defer stream.Close(q.db.Context())

	processedUntil := nowFunc()
	// process unprocessed tasks scheduled before we started watching
	for {
		task, err := q.GetNext(topic)
		if err != nil {
			return err
		}

		if task == nil {
			break
		}

		processedUntil = task.Meta.Created
		cb(*task)
	}

	for stream.Next(q.db.Context()) {
		var evt event

		if err := stream.Decode(&evt); err != nil {
			continue
		}

		// already processed
		if evt.Task.Meta.Created.Before(processedUntil) {
			continue
		}

		task, err := q.GetNextById(evt.Task.Id)
		if err != nil {
			_ = q.Err(evt.Task.Id.Hex(), err)
			continue
		}

		if task != nil {
			cb(*task)
		}
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
			"meta.completed": nowFunc(),
		}})
}

func (q *Queue) Err(id string, err error) error {
	oId, e := primitive.ObjectIDFromHex(id)
	if e != nil {
		return e
	}

	return q.db.UpdateOne(
		bson.M{"_id": oId},
		bson.M{"$set": bson.M{
			"state":          StateError,
			"meta.completed": nowFunc(),
			"message":        err.Error()},
		})
}

func (q *Queue) Selfcare(topic string, timeout time.Duration) error {
	// re-schedule long-running tasks
	// this only happens if the processor could not ack the task, i.e. the application crashed

	if timeout == 0 {
		timeout = DefaultTimeout
	}

	query := bson.M{
		"state":           StateRunning,
		"meta.dispatched": bson.M{"$lt": nowFunc().Add(timeout)},
	}
	if len(topic) > 0 {
		query["topic"] = topic
	}

	err1 := q.db.UpdateMany(
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
	if len(topic) > 0 {
		query["topic"] = topic
	}

	err2 := q.db.UpdateMany(
		query,
		bson.M{"$set": bson.M{
			"state":          StateError,
			"meta.completed": nowFunc()},
		})

	if err1 != nil {
		return err1
	}

	if err2 != nil {
		return err2
	}

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
