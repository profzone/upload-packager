package packages

import "github.com/profzone/eden-framework/pkg/courier"

var Router = courier.NewRouter(PackageGroup{})

type PackageGroup struct {
	courier.EmptyOperator
}

func (PackageGroup) Path() string {
	return "/packages"
}
