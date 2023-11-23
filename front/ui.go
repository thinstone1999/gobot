package front

import "github.com/zserge/lorca"

// 用户界面
type UI struct {
	lorca.UI
	silent bool
	logId  int64
}

func NewUI(height, width int) *UI {
	args := make([]string, 0, 4)
	args = append(args, "--remote-allow-origins=*")
	args = append(args, "--no-sandbox")
	gui, er := lorca.New("", "", width, height, args...)
	if er != nil {
		panic(er)
	}

	// 注册函数给js
	// js中全局变量GO 即等于FromJS
	gui.Bind("GO", FromJS)
	ui := &UI{
		UI: gui,
	}
	return ui
}

func (app *UI) Done() <-chan struct{} {
	return app.UI.Done()
}

func (app *UI) SetSilent(s bool) {
	app.silent = s
}

func (app *UI) GetSilent() bool {
	return app.silent
}
