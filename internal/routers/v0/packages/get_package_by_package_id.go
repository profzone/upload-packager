package packages

import (
	"context"
	"github.com/profzone/eden-framework/pkg/courier"
	"github.com/profzone/eden-framework/pkg/courier/httpx"
	"github.com/profzone/eden-framework/pkg/enumeration"
)

func init() {
	Router.Register(courier.NewRouter(GetPackageByPackageID{}))
}

// 通过ID获取压缩包
type GetPackageByPackageID struct {
	httpx.MethodGet
	// 压缩包ID
	PackageID uint64 `name:"packageId,string" in:"path"`
	// 是否返回文件内容
	WithContent enumeration.Bool `name:"withContent" in:"query" default:"FALSE"`
}

func (req GetPackageByPackageID) Path() string {
	return "/:packageId"
}

func (req GetPackageByPackageID) Output(ctx context.Context) (result interface{}, err error) {
	return
}
