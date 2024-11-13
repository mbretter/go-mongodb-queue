[![](https://github.com/mbretter/go-mongodb-queue/actions/workflows/test.yml/badge.svg)](https://github.com/mbretter/go-mongodb-queue/actions/workflows/test.yml)
[![](https://goreportcard.com/badge/mbretter/go-mongodb-queue)](https://goreportcard.com/report/mbretter/go-mongodb-queue "Go Report Card")
[![codecov](https://codecov.io/gh/mbretter/go-mongodb-queue/graph/badge.svg?token=YMBMKY7W9X)](https://codecov.io/gh/mbretter/go-mongodb-queue)
[![GoDoc](https://godoc.org/github.com/mbretter/go-mongodb-queue?status.svg)](https://pkg.go.dev/github.com/mbretter/go-mongodb-queue)

This is a dead simple queuing system based on MongoDB. 
It is primarily build upon MongoDB's [change streams](https://www.mongodb.com/docs/manual/changeStreams/), this provides 
the possibility to use an event based system, instead of using a polling approach.

MongoDB change-streams are available, if you have configured a replica-set, as a fallback this packages supports 
polling too.

The motivation was to build an easy-to-integrate queuing system without sophisticated features, without external 
dependencies, and with direct integration into your application.

## Install

```
go get github.com/mbretter/go-mongodb-queue
```

import

```go
import queue "github.com/mbretter/go-mongodb-queue"
```

## Features

There are not that many, it supports retries until a maximum number of tries have been reached, and it has a 
default timeout for tasks, which is set to 5 minutes, if running the selfcare function.

Along the task, any arbitrary data can be stored.

Each task belongs to a topic, when publishing to a topic, the handler of this topic gets the first unprocessed task.

You can use either the event based `Subscribe` function, or the `GetNext` function which is needed for polling.
It both cases it is totally safe to have multiple consumers running on the same topic.


```go
ctx := context.TODO()
// connect to the mongo database using the mongo-driver
// mongodbUri contains the uri to your mongodb instance
client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbUri))
if err != nil {
    log.Fatal(err)
}
defer client.Disconnect(ctx)

// get database and collection
collection := client.Database("mydb").Collection("queue")

// make the queue-db
queueDb := queue.NewStdDb(collection, ctx)

// make the queue 
qu := queue.NewQueue(queueDb)
```

## Publish

You can publish to any topic, the topic acts like a filter for your tasks, the payload can be any arbitrary data.

```go
type Payload struct {
    Name string `bson:"name"`
    Desc string `bson:"desc"`
    Num  int    `bson:"num"`
}

payload := Payload{
    Name: "Arnold Schwarzenegger",
    Desc: "I'll be back",
    Num:  73,
}

task, err := qu.Publish("some.topic", &payload)
if err != nil {
    log.Fatal(err)
}
```

## Subscribe

Any handler/application can subscribe to a certain topic, the callback function receives a copy of the task.

After processing the task you have to `Ack` it, or mark it as `Err`.

Here is a small snippet which demonstrates the usage of subscribe using goroutines.
```go
// define your worker function
workerFunc := func(qu *queue.Queue, task queue.Task) {
    fmt.Println("worker", task)
	// after processing the task you have to acknowledge it
    _ = qu.Ack(task.Id.Hex())
}

var wg sync.WaitGroup
// subscribe and pass the worker function
err := qu.Subscribe("some.topic", func(t queue.Task) {
    wg.Add(1)
    go func() {
        defer wg.Done()
        workerFunc(qu, t)
    }()
})

if err != nil {
    log.Fatal(err)
}

wg.Wait()
```

On startup, the `Subscribe` function checks for unprocessed tasks scheduled before we subscribed, because existing 
tasks will not be covered by the MongoDB change-stream.

## Ack/Err

After processing a task you have to acknowledge, that you have processed the task by using `Ack`.
In case of an error you can use the `Err` function to mark the task as failed.

```go
err := qu.Ack(task.Id.Hex())
if err != nil {
    log.Fatal(err)
}

qu.Err(task.Id.Hex(), errors.New("something went wrong"))
```

## Polling

You have to loop over `GetNext`, `GetNext` returns a nil task, if no unprocessed task was found or the topic.
It is safe to use `GetNext` for the same topic from different processes, there will be no race conditions, because MongoDB's atomic 
`FindOneAndUpdate` operation is used.

```go
for {
    task, err := qu.GetNext("some.topic")
    if err != nil {
        log.Fatal(err)
    }
    
    if task == nil {
        time.Sleep(time.Millisecond * 100)
    } else {
        // process the task
        _ = qu.Ack(task.Id.Hex())
    }
}
```

## Reschedule

If a task had an error, and you want to process this task again, you can use `Reschedule`.

```go
newTask, err := Reschedule(task)
```

When rescheduling, the original task remains untouched, there will create a new task with the same payload and the 
initial tries value will be increased. 

## Selfcare

The selfcare function re-schedules long-running tasks, this might happen, if the application could not acknowledge 
the task, and it sets the task to the error state, if the maximum number of tries have been exceeded.

The selfcare function might be run per topic, if no topic was given, the selfcare runs over all topics.
As second argument, the timeout for long-running tasks can be given, if no timeout was given it defaults to 5 minutes.

## CreateIndexes

On first use, you have to call this function, or set the indexes manually on the queue collection.
There will be created two indexes, one on `topic` and `state`, the other one is a TTL-index, which removes completed 
tasks after one hour. 

## mongodb docker-compose

This is a config sample for bringing up a MongoDB-replicaset for testing, you have to use fixed ip-addresses for your containers, 
otherwise you will get connection problems, because they are sent back to the client, when connecting. 

```yaml
services:
  mongo1:
    hostname: mongo1
    image: mongo:8
    ports:
      - 27027:27017
    restart: always
    volumes:
      - /data/mongodb-rs1:/data/db
    command: mongod --replSet rs0
    networks:
      customnetwork:
        ipv4_address: 172.20.0.3
  mongo2:
    hostname: mongo2
    image: mongo:8
    ports:
      - 27028:27017
    restart: always
    volumes:
      - /data/mongodb-rs2:/data/db
    command: mongod --replSet rs0
    networks:
      customnetwork:
        ipv4_address: 172.20.0.4
  mongo3:
    hostname: mongo3
    image: mongo:8
    ports:
      - 27029:27017
    restart: always
    volumes:
      - /data/mongodb-rs3:/data/db
    command: mongod --replSet rs0
    networks:
      customnetwork:
        ipv4_address: 172.20.0.5

networks:
  customnetwork:
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

After bringing up the containers the first time, you have to initialize the replicaset:
```shell
docker compose exec mongo1 mongosh
> rs.initiate({ _id: "rs0", members: [ { _id: 0, host: "172.20.0.3:27017" }, { _id: 1, host: "172.20.0.4:27017" }, { _id: 2, host: "172.20.0.5:27017" }] })
```