package main

import (
	"context"
	"flag"
	"fmt"
	queue "github.com/mbretter/go-mongodb-queue"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"sync"
)

type Payload struct {
	Name string `bson:"name"`
	Desc string `bson:"desc"`
	Num  int    `bson:"num"`
}

func main() {
	var mongodbUri = flag.String("u", "", "mongodb url")
	var dbName = flag.String("d", "", "mongodb database name")
	var collName = flag.String("c", "queue", "mongodb collection name")
	var publish = flag.String("p", "", "publish topic")
	var getnext = flag.String("g", "", "next topic")
	var ackId = flag.String("a", "", "ack id")
	var selfcare = flag.Bool("sc", false, "run selfcare")
	var createIndexes = flag.Bool("i", false, "create indexes")
	var subscribe = flag.String("s", "", "subscribe on topic")
	flag.Parse()

	if len(*mongodbUri) == 0 {
		log.Fatal("mongodb uri missing")
	}

	if len(*dbName) == 0 {
		log.Fatal("mongodb database name missing")
	}

	ctx := context.TODO()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(*mongodbUri))
	if err != nil {
		log.Fatal(err)
	}
	//goland:noinspection ALL
	defer client.Disconnect(ctx)

	collection := client.Database(*dbName).Collection(*collName)

	queueDb := queue.NewStdDb(collection, ctx)
	qu := queue.NewQueue(queueDb)

	payload := Payload{
		Name: "Arnold Schwarzenegger",
		Desc: "I'll be back",
		Num:  73,
	}

	if *subscribe != "" {
		// inlined to be more readable, practically this func would be somewhere else
		workerFunc := func(qu *queue.Queue, task queue.Task) {
			fmt.Println("worker", task)
			_ = qu.Ack(task.Id.Hex())
		}

		var wg sync.WaitGroup
		err := qu.Subscribe(*subscribe, func(t queue.Task) {
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
	}

	if *publish != "" {
		task, err := qu.Publish(*publish, &payload, queue.DefaultMaxTries)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(*task)
	}

	if *getnext != "" {
		task, err := qu.GetNext(*getnext)
		if err != nil {
			log.Fatal(err)
		}

		if task != nil {
			fmt.Println(*task)
		}
	}

	if *ackId != "" {
		err := qu.Ack(*ackId)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *selfcare {
		err := qu.Selfcare(nil)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *createIndexes {
		err := qu.CreateIndexes()
		if err != nil {
			log.Fatal(err)
		}
	}
}
