package v0

import (
	"context"
	"github.com/profzone/eden-framework/pkg/courier"
	"github.com/profzone/eden-framework/pkg/courier/httpx"
	"longhorn/upload-packager/internal/global"
)

func init() {
	Router.Register(courier.NewRouter(GetOp{}))
}

//
type GetOp struct {
	httpx.MethodGet
}

func (req GetOp) Path() string {
	return ""
}

func (req GetOp) Output(ctx context.Context) (result interface{}, err error) {
	value, err := global.Config.Raft.Pop()
	if err != nil {
		return
	}

	return value, nil
}
