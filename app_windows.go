package gobot

import (
	"github.com/Gonewithmyself/gobot/front"
)

func (app *App) Init(opt *Options) {
	app.BaseApp.Init(opt)
	app.UI = front.NewUI(Height, Width)
	app.Fs = front.NewFileServer(opt.StaticDir)
	app.Os = OsWindows
}

func (app *App) getUI() *front.UI {
	return app.UI
}

// func (app *App) PrintStressStatus() {
// 	app.UI.Print("hello", map[string]interface{}{
// 		"go":     1,
// 		"python": 99,
// 	})
// }
