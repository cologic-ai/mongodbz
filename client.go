package mongodbz

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	_ "go.mongodb.org/mongo-driver/mongo/readpref"
	"reflect"
	"time"
)

func New(config Config) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cli, err := mongo.Connect(ctx, options.Client().ApplyURI(config.URL), options.Client())
	if err != nil {
		return nil, err
	}

	err = cli.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}
	mcli := &MongoClient{
		cli:             cli,
		DB:              cli.Database(config.Name),
		handlersMap:     map[string]MongoHandler{},
		closeChan:       make(chan struct{}),
		numberOfWorkers: config.NumberOfJobs,
	}

	return mcli, nil
}

func (m *MongoClient) AddWorkItem(item any, callback func(any, error)) {
	m.workItem <- WorkItem{
		Item:     item,
		Callback: callback,
	}
}

type syncWorkItem struct {
	item any
	err  error
}

func (m *MongoClient) AddWorkItemSync(item any) (any, error) {
	callback := make(chan syncWorkItem)
	m.workItem <- WorkItem{
		Item: item,
		Callback: func(a any, err error) {
			callback <- syncWorkItem{a, err}
		},
	}
	result := <-callback
	return result.item, result.err
}

func (m *MongoClient) AddHandler(handler MongoHandler) {
	m.handlers = append(m.handlers, handler)
	m.handlersMap[handler.InputType()] = handler
}

func (m *MongoClient) FindHandler(item any) MongoHandler {
	h, ok := m.handlersMap[FormatType(reflect.TypeOf(item))]
	if !ok {
		return nil
	}
	return h
}

func (m *MongoClient) Start() {
	m.workItem = make(chan WorkItem, 100)
	m.wg.Add(m.numberOfWorkers)
	for i := 0; i < m.numberOfWorkers; i++ {
		go m.worker()
	}
}

func (m *MongoClient) worker() {
	defer m.wg.Done()
	for {
		select {
		case <-m.closeChan:
			return
		case item := <-m.workItem:
			handler := m.FindHandler(item.Item)
			if handler == nil {
				if item.Callback != nil {
					item.Callback(nil, fmt.Errorf("no handler found for %s", FormatType(reflect.TypeOf(item.Item))))
				}
				continue
			}
			handler.Process(m, item.Item, item.Callback)
		}
	}
}

func (m *MongoClient) Close() {
	close(m.closeChan)
	m.wg.Wait()
}
