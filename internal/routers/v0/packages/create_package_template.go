package packages

import (
	"context"
	"github.com/profzone/eden-framework/pkg/courier"
	"github.com/profzone/eden-framework/pkg/courier/httpx"
)

func init() {
	Router.Register(courier.NewRouter(CreatePackageTemplate{}))
}

// 创建导出任务
type CreatePackageTemplate struct {
	httpx.MethodPost
}

func (req CreatePackageTemplate) Path() string {
	return "/template"
}

func (req CreatePackageTemplate) Output(ctx context.Context) (result interface{}, err error) {
	return
}
