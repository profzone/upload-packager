package v0

import (
	"github.com/profzone/eden-framework/pkg/courier"
	"longhorn/upload-packager/internal/routers/v0/packages"
	"longhorn/upload-packager/internal/routers/v0/peers"
	"longhorn/upload-packager/internal/routers/v0/templates"
)

var Router = courier.NewRouter(V0Group{})

func init() {
	Router.Register(packages.Router)
	Router.Register(templates.Router)
	Router.Register(peers.Router)
}

type V0Group struct {
	courier.EmptyOperator
}

func (V0Group) Path() string {
	return "/v0"
}
