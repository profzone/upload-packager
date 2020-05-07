package main

import (
	"github.com/profzone/eden-framework/pkg/application"
	"github.com/sirupsen/logrus"
	"longhorn/upload-packager/internal/global"
	"longhorn/upload-packager/internal/routers"
)

func main() {
	app := application.NewApplication(runner, &global.Config)
	go app.Start()
	app.WaitStop(func() error {
		return nil
	})
}

func runner(app *application.Application) error {
	logrus.SetLevel(global.Config.LogLevel)
	global.Config.Raft.Init()
	go global.Config.GRPCServer.Serve(routers.RootRouter)
	return global.Config.HTTPServer.Serve(routers.RootRouter)
}
