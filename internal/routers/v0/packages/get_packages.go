package packages

import (
	"context"
	"github.com/profzone/eden-framework/pkg/courier"
	"github.com/profzone/eden-framework/pkg/courier/httpx"
)

func init() {
	Router.Register(courier.NewRouter(GetPackages{}))
}

// 获取压缩包列表
type GetPackages struct {
	httpx.MethodGet
}

func (req GetPackages) Path() string {
	return ""
}

func (req GetPackages) Output(ctx context.Context) (result interface{}, err error) {
	return
}
