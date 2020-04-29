package routers

import (
	"github.com/profzone/eden-framework/pkg/courier"
	"github.com/profzone/eden-framework/pkg/courier/swagger"
	v0 "longhorn/upload-packager/internal/routers/v0"
)

var RootRouter = courier.NewRouter(RootGroup{})

func init() {
	RootRouter.Register(swagger.SwaggerRouter)
	RootRouter.Register(v0.Router)
}

type RootGroup struct {
	courier.EmptyOperator
}

func (RootGroup) Path() string {
	return "/packager"
}
