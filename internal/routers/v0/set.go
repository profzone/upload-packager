package v0

import (
	"context"
	"github.com/profzone/eden-framework/pkg/courier"
	"github.com/profzone/eden-framework/pkg/courier/httpx"
	"longhorn/upload-packager/internal/global"
)

func init() {
	Router.Register(courier.NewRouter(SetOp{}))
}

//
type SetOp struct {
	httpx.MethodPost
	Value string `json:"value" in:"query"`
}

func (req SetOp) Path() string {
	return ""
}

func (req SetOp) Output(ctx context.Context) (result interface{}, err error) {
	err = global.Config.Raft.Push([]byte(req.Value))
	return
}
