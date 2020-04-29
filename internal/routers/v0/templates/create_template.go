package templates

import (
	"context"
	"github.com/profzone/eden-framework/pkg/courier"
	"github.com/profzone/eden-framework/pkg/courier/httpx"
)

func init() {
	Router.Register(courier.NewRouter(CreateTemplate{}))
}

// 创建模板
type CreateTemplate struct {
	httpx.MethodPost
}

func (req CreateTemplate) Path() string {
	return ""
}

func (req CreateTemplate) Output(ctx context.Context) (result interface{}, err error) {
	return
}
