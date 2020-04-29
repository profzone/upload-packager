package packages

import (
	"context"
	"github.com/profzone/eden-framework/pkg/courier"
	"github.com/profzone/eden-framework/pkg/courier/httpx"
)

func init() {
	Router.Register(courier.NewRouter(CreatePackageExport{}))
}

// 创建导出任务
type CreatePackageExport struct {
	httpx.MethodPost
}

func (req CreatePackageExport) Path() string {
	return "/export"
}

func (req CreatePackageExport) Output(ctx context.Context) (result interface{}, err error) {
	return
}
