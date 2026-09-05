package queue

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Queue interface {
	Publish(ctx context.Context, topic string, payload any, opts ...*PublishOptions) (Task, error)
	GetNext(ctx context.Context, topic string) (Task, error)
	GetNextById(ctx context.Context, id bson.ObjectID) (Task, error)
	Reschedule(ctx context.Context, task Task) (Task, error)
	Subscribe(ctx context.Context, topic string, cb Callback) error
	Ack(ctx context.Context, id string) error
	Err(ctx context.Context, id string, err error) error
	Selfcare(ctx context.Context, topic string, timeout time.Duration) error
	CreateIndexes(ctx context.Context) error
}

type DbInterface interface {
	InsertOne(ctx context.Context, document any) (bson.ObjectID, error)
	FindOneAndUpdate(ctx context.Context, filter any, update any, opts ...options.Lister[options.FindOneAndUpdateOptions]) *mongo.SingleResult
	UpdateOne(ctx context.Context, filter any, update any) error
	UpdateMany(ctx context.Context, filter any, update any) error
	Watch(ctx context.Context, pipeline any) (ChangeStreamInterface, error)
	CreateIndexes(ctx context.Context, index []mongo.IndexModel) error
}

type queue struct {
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

type Task interface {
	GetId() bson.ObjectID
	GetTopic() string
	GetPayload() any
	GetTries() uint
	GetMaxTries() uint
	GetState() string
	GetMessage() string
	GetMeta() Meta
}

type task struct {
	Id       bson.ObjectID `bson:"_id,omitempty"`
	Topic    string        `bson:"topic"`
	Payload  any           `bson:"payload"`
	Tries    uint          `bson:"tries"`
	MaxTries uint          `bson:"maxtries"`
	State    string        `bson:"state"`
	Message  string        `bson:"message"`
	Meta     Meta
}

func (t *task) GetId() bson.ObjectID {
	return t.Id
}

func (t *task) GetTopic() string {
	return t.Topic
}

func (t *task) GetPayload() any {
	return t.Payload
}

func (t *task) GetTries() uint {
	return t.Tries
}

func (t *task) GetMaxTries() uint {
	return t.MaxTries
}

func (t *task) GetState() string {
	return t.State
}

func (t *task) GetMessage() string {
	return t.Message
}

func (t *task) GetMeta() Meta {
	return t.Meta
}

type event struct {
	Task Task `bson:"fullDocument"`
}

// NewQueue initializes a new Queue instance with the provided DbInterface.
func NewQueue(db DbInterface) Queue {
	queue := queue{
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

// NewPublishOptions creates a new PublishOptions with default settings.
func NewPublishOptions() *PublishOptions {
	return &PublishOptions{
		MaxTries: 0,
		Tries:    -1,
	}
}

// SetMaxTries sets the maximum number of retry attempts for publishing. Returns the updated PublishOptions instance.
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
func (q *queue) Publish(ctx context.Context, topic string, payload any, opts ...*PublishOptions) (Task, error) {

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

	t := task{
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

	insertedId, err := q.db.InsertOne(ctx, t)
	if err != nil {
		return nil, err
	}

	t.Id = insertedId

	return &t, nil
}

// GetNext retrieves the next item from the queue for the given topic, marks it as running, and increments its tries count.
func (q *queue) GetNext(ctx context.Context, topic string) (Task, error) {
	t := task{}
	res := q.db.FindOneAndUpdate(ctx, bson.M{
		"topic": topic,
		"state": StatePending,
		"$expr": bson.M{"$lt": bson.A{"$tries", "$maxtries"}},
	},
		bson.M{
			"$set": bson.M{"state": StateRunning, "meta.dispatched": nowFunc()},
			"$inc": bson.M{"tries": 1},
		},
		options.FindOneAndUpdate().SetSort(bson.M{"meta.scheduled": 1}).SetReturnDocument(options.After),
	)

	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return nil, nil
	}

	if err := res.Decode(&t); err != nil {
		return nil, err
	}

	return &t, nil
}

// GetNextById retrieves the next pending task by its ID, transitions it to the running state, and increments its tries count.
func (q *queue) GetNextById(ctx context.Context, id bson.ObjectID) (Task, error) {
	t := task{}
	res := q.db.FindOneAndUpdate(ctx, bson.M{
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

// Reschedule republishes a task to the queue, retaining its topic, payload, tries, and maxTries settings.
func (q *queue) Reschedule(ctx context.Context, task Task) (Task, error) {
	return q.Publish(ctx, task.GetTopic(), task.GetPayload(), NewPublishOptions().setTries(task.GetTries()).SetMaxTries(task.GetMaxTries()))
}

type Callback func(t Task)

// Subscribe listens for new tasks on a given topic and calls the provided callback when a new task is available.
// It processes unprocessed tasks scheduled before starting the watch and continuously monitors for new tasks.
func (q *queue) Subscribe(ctx context.Context, topic string, cb Callback) error {
	pipeline := bson.D{{Key: "$match", Value: bson.D{
		{Key: "operationType", Value: "insert"},
		{Key: "fullDocument.topic", Value: topic},
		{Key: "fullDocument.state", Value: StatePending}}},
	}

	stream, err := q.db.Watch(ctx, mongo.Pipeline{pipeline})
	if err != nil {
		return err
	}

	defer stream.Close(ctx)

	processedUntil := nowFunc()
	// process unprocessed tasks scheduled before we started watching
	for {
		task, err := q.GetNext(ctx, topic)
		if err != nil {
			return err
		}

		if task == nil {
			break
		}

		processedUntil = task.GetMeta().Created
		cb(task)
	}

	for stream.Next(ctx) {
		var evt event

		if err := stream.Decode(&evt); err != nil {
			continue
		}

		// already processed
		if evt.Task.GetMeta().Created.Before(processedUntil) {
			continue
		}

		task, err := q.GetNextById(ctx, evt.Task.GetId())
		if err != nil {
			_ = q.Err(ctx, evt.Task.GetId().Hex(), err)
			continue
		}

		if task != nil {
			cb(task)
		}
	}

	return nil
}

// Ack acknowledges a task completion by its ID, updating its state to "completed" and setting the completion timestamp.
func (q *queue) Ack(ctx context.Context, id string) error {
	oId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	return q.db.UpdateOne(
		ctx,
		bson.M{"_id": oId},
		bson.M{"$set": bson.M{
			"state":          StateCompleted,
			"meta.completed": nowFunc(),
		}})
}

// Err updates the state of a task to "error" by its ID, setting the completion time and storing the error message.
func (q *queue) Err(ctx context.Context, id string, err error) error {
	oId, e := bson.ObjectIDFromHex(id)
	if e != nil {
		return e
	}

	return q.db.UpdateOne(
		ctx,
		bson.M{"_id": oId},
		bson.M{"$set": bson.M{
			"state":          StateError,
			"meta.completed": nowFunc(),
			"message":        err.Error()},
		})
}

// Selfcare re-schedules long-running tasks and sets tasks exceeding max tries to error state.
// It updates tasks in an ongoing state that haven't been acknowledged within a specific timeout period.
// If timeout is zero, the default timeout value is used. Optionally, tasks can be filtered by topic.
func (q *queue) Selfcare(ctx context.Context, topic string, timeout time.Duration) error {
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
		ctx,
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
		ctx,
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

// CreateIndexes creates MongoDB indexes for the task collection to improve query performance and manage TTL for completed tasks.
func (q *queue) CreateIndexes(ctx context.Context) error {
	err := q.db.CreateIndexes(ctx, []mongo.IndexModel{{
		Keys: bson.D{{Key: "topic", Value: 1}, {Key: "state", Value: 1}},
	}, {
		Keys: bson.D{{Key: "meta.completed", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(3600),
	}})

	return err
}
