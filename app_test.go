package odogoron

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"testing"
)

type mockEventSource struct {
	ch chan EventCtx
}

func (mes mockEventSource) GetChannel() chan EventCtx {
	return mes.ch
}

func (mes mockEventSource) GetMapper() func(context.Context, interface{}) (interface{}, error) {
	return func(_ context.Context, event interface{}) (interface{}, error) {
		eventCtx, ok := event.(*EventCtx)
		if !ok {
			return nil, errors.New("mockEventSource: error map event")
		}
		return *eventCtx, nil
	}
}

func (mes mockEventSource) Shutdown() {
}

func (mes mockEventSource) SendEvent(ctx EventCtx) {
	mes.ch <- ctx
}

type mockEventCtx struct {
	eventName string
}

func (m mockEventCtx) GetName() string {
	return m.eventName
}

func (m mockEventCtx) GetBody() any {
	return "Body of CTX"
}

func getMockEventSource() mockEventSource {
	eventsChan := make(chan EventCtx)
	return mockEventSource{
		ch: eventsChan,
	}
}

func TestNewApp(t *testing.T) {
	app := New(getMockEventSource())
	assert.False(t, app.isRunning)
}

func TestRunApp(t *testing.T) {
	m := getMockEventSource()
	e := mockEventCtx{eventName: "testingEvent"}
	go m.SendEvent(e)

	app := New(m)

	h := func(ctx EventCtx) {
		assert.Equal(t, ctx, e)
		app.Shutdown()
	}

	app.RegisterHandler("testingEvent", h)
	app.Run()
}
