package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	queue "github.com/mbretter/go-mongodb-queue/v3"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Payload struct {
	Name string `bson:"name"`
	Desc string `bson:"desc"`
	Num  int    `bson:"num"`
}

func main() {
	os.Chdir("../..")
	var collName = flag.String("c", "queue", "mongodb collection name")
	var publish = flag.String("p", "", "publish topic")
	var getnext = flag.String("g", "", "next topic")
	var ackId = flag.String("a", "", "ack id")
	var selfcare = flag.Bool("sc", false, "run selfcare")
	var createIndexes = flag.Bool("i", false, "create indexes")
	var subscribe = flag.String("s", "", "subscribe on topic")
	flag.Parse()

	mongodbUri := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_DB")
	if len(mongodbUri) == 0 {
		log.Fatal("mongodb uri missing")
	}

	if len(dbName) == 0 {
		log.Fatal("mongodb database name missing")
	}

	ctx := context.TODO()
	client, err := mongo.Connect(options.Client().ApplyURI(mongodbUri))
	if err != nil {
		log.Fatal(err)
	}
	//goland:noinspection ALL
	defer client.Disconnect(ctx)

	collection := client.Database(dbName).Collection(*collName)

	queueDb := queue.NewStdDb(collection)
	qu := queue.NewQueue(queueDb)

	payload := Payload{
		Name: "Arnold Schwarzenegger",
		Desc: "I'll be back",
		Num:  73,
	}

	if *subscribe != "" {
		// inlined to be more readable, practically this func would be somewhere else
		workerFunc := func(qu queue.Queue, task queue.Task) {
			fmt.Println("worker", task)
			_ = qu.Ack(ctx, task.GetId().Hex())
		}

		var wg sync.WaitGroup
		err := qu.Subscribe(ctx, *subscribe, func(t queue.Task) {
			wg.Go(func() {
				workerFunc(qu, t)
			})
		})

		if err != nil {
			log.Fatal(err)
		}

		wg.Wait()
	}

	if *publish != "" {
		opts := queue.NewPublishOptions().SetMaxTries(1)
		task, err := qu.Publish(ctx, *publish, &payload, opts)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(task)
	}

	if *getnext != "" {
		task, err := qu.GetNext(ctx, *getnext)
		if err != nil {
			log.Fatal(err)
		}

		if task != nil {
			fmt.Println(task)
		}
	}

	if *ackId != "" {
		err := qu.Ack(ctx, *ackId)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *selfcare {
		err := qu.Selfcare(ctx, "", 0)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *createIndexes {
		err := qu.CreateIndexes(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
}
