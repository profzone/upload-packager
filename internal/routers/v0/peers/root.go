package peers

import "github.com/profzone/eden-framework/pkg/courier"

var Router = courier.NewRouter(PeerGroup{})

type PeerGroup struct {
	courier.EmptyOperator
}

func (PeerGroup) Path() string {
	return "/peers"
}
