package odogoron

import (
	"github.com/reactivex/rxgo/v2"
	"golang.org/x/net/context"
	"log"
	"sync"
)

const DefaultConcurrency = 256 * 1024

type App struct {
	mutex       sync.Mutex
	isRunning   bool
	config      Config
	eventSource EventSource
	handlers    handlersHub
	shutdownCh  chan struct{}
}

type handlersHub struct {
	list map[string]HandlerFunction
}

type EventSource interface {
	GetChannel() chan EventCtx
	GetMapper() func(context.Context, interface{}) (interface{}, error)
	Shutdown()
}

type EventCtx interface {
	GetName() string
	GetBody() any
}

type Config struct {
	AppName      string
	ErrorHandler ErrorHandler
	Concurrency  int
}

func New(source EventSource, config ...Config) *App {
	app := &App{
		isRunning: false,
		handlers: handlersHub{
			list: map[string]HandlerFunction{},
		},
		config:      Config{},
		eventSource: source,
	}

	// Override config if provided
	if len(config) > 0 {
		app.config = config[0]
	}

	if app.config.ErrorHandler == nil {
		app.config.ErrorHandler = DefaultErrorHandler
	}

	if app.config.Concurrency <= 0 {
		app.config.Concurrency = DefaultConcurrency
	}

	app.shutdownCh = make(chan struct{})

	return app
}

func (app *App) Run() {
	listener := rxgo.Defer([]rxgo.Producer{
		func(ctx context.Context, next chan<- rxgo.Item) {
			eventChan := app.eventSource.GetChannel()
		getterLoop:
			for {
				select {
				case <-app.shutdownCh:
					app.eventSource.Shutdown()
					app.isRunning = false
					break getterLoop

				case event := <-eventChan:
					next <- rxgo.Item{V: &event}
				}
			}
		},
	})

	mapFunc := app.eventSource.GetMapper()
	<-listener.Map(mapFunc, rxgo.WithPool(app.config.Concurrency)).
		ForEach(
			func(item interface{}) {
				eventCtx, ok := item.(EventCtx)
				if !ok {
					log.Fatal("Odogoron: not EventCtx interface")
				}

				defer recoverHandling(app, eventCtx)
				handler, ok := app.handlers.list[eventCtx.GetName()]
				if !ok {
					log.Println("Odogoron: event handler not found:", eventCtx.GetName())
					return
				}
				handler(eventCtx)
			},
			func(err error) {
				log.Println(err)
			},
			func() {})
}

type ErrorHandler func(error, EventCtx)

var DefaultErrorHandler = func(err error, event EventCtx) {
	log.Println(err)
}

func recoverHandling(app *App, eventCtx EventCtx) {
	if e := recover(); e != nil {
		app.config.ErrorHandler(e.(error), eventCtx)
	}
}

func (app *App) Shutdown() {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	app.shutdownCh <- struct{}{}
	close(app.shutdownCh)
}
