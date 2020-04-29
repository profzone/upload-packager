package templates

import "github.com/profzone/eden-framework/pkg/courier"

var Router = courier.NewRouter(TemplateGroup{})

type TemplateGroup struct {
	courier.EmptyOperator
}

func (TemplateGroup) Path() string {
	return "/templates"
}
