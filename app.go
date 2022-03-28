package odogoron

type App struct {
	isRunning bool
}

func New() *App {
	app := &App{
		isRunning: false,
	}

	return app
}
