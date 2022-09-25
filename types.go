package mongodbz

import (
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
)

type Handler[T, U any] struct {
	TType          string
	UType          string
	closeChan      chan struct{}
	f              func(*MongoClient, T) (U, error)
	collectionName string
}

type MongoHandler interface {
	InputType() string
	OutputType() string
	Process(cli *MongoClient, item any, callback func(any, error))
	ProcessSync(cli *MongoClient, item any) (any, error)
}

type WorkItem struct {
	Item     any
	Callback func(any, error)
}

type Conifg struct {
	//MongoDB URL, ignore if Client is set.
	URL string
	//The Name of the Database
	Name string
	//Number of workers
	NumberOfJobs int

	//Mongo client, If there is any Auth or TLS configure require, set Client
	Client *mongo.Client
}

type MongoClient struct {
	cli         *mongo.Client
	DB          *mongo.Database
	handlers    []MongoHandler
	handlersMap map[string]MongoHandler

	closeChan       chan struct{}
	numberOfWorkers int
	workItem        chan WorkItem
	wg              sync.WaitGroup
}
