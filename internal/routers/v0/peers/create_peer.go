package peers

import (
	"context"
	"github.com/profzone/eden-framework/pkg/courier"
	"github.com/profzone/eden-framework/pkg/courier/httpx"
	"longhorn/upload-packager/internal/global"
	"longhorn/upload-packager/pkg/consensus"
)

func init() {
	Router.Register(courier.NewRouter(CreatePeer{}))
}

// 创建一个集群节点
type CreatePeer struct {
	httpx.MethodPost
	Body consensus.CreatePeerBody `name:"body" in:"body"`
}

func (req CreatePeer) Path() string {
	return ""
}

func (req CreatePeer) Output(ctx context.Context) (result interface{}, err error) {
	future := global.Config.Raft.AddVoter(req.Body.PeerID, req.Body.PeerAddr, 0, 0)
	err = future.Error()
	return
}
