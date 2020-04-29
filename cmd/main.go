package main

import (
	"github.com/profzone/eden-framework/pkg/application"
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
	global.Config.Raft.Init()
	go global.Config.GRPCServer.Serve(routers.RootRouter)
	return global.Config.HTTPServer.Serve(routers.RootRouter)
}
