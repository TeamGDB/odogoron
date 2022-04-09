package odogoron

type HandlerFunction func(ctx EventCtx)

func (app *App) RegisterHandler(eventName string, function HandlerFunction) {
	app.mutex.Lock()
	app.handlers.list[eventName] = function
	app.mutex.Unlock()
}
