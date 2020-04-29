package packages

import (
	"context"
	"github.com/profzone/eden-framework/pkg/courier"
	"github.com/profzone/eden-framework/pkg/courier/httpx"
)

func init() {
	Router.Register(courier.NewRouter(CreatePackageArchive{}))
}

// 创建压缩包任务
type CreatePackageArchive struct {
	httpx.MethodPost
}

func (req CreatePackageArchive) Path() string {
	return "/archive"
}

type CreatePackageResp struct {
	ID uint64 `json:"id,string"`
}

func (req CreatePackageArchive) Output(ctx context.Context) (result interface{}, err error) {
	return
}
