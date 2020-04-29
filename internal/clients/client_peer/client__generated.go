package client_peer

import (
	"fmt"
	git_chinawayltd_com_golib_tools_courier "github.com/profzone/eden-framework/pkg/courier"
	git_chinawayltd_com_golib_tools_courier_client "github.com/profzone/eden-framework/pkg/courier/client"
	git_chinawayltd_com_golib_tools_courier_status_error "github.com/profzone/eden-framework/pkg/courier/status_error"
)

type ClientIDInterface interface {
	CreatePeer(metas ...git_chinawayltd_com_golib_tools_courier.Metadata) (resp *CreatePeerResponse, err error)
}

type ClientPeer struct {
	git_chinawayltd_com_golib_tools_courier_client.Client
}

func (ClientPeer) MarshalDefaults(v interface{}) {
	if cl, ok := v.(*ClientPeer); ok {
		cl.Name = "id"
		cl.Client.MarshalDefaults(&cl.Client)
	}
}

func (c ClientPeer) Init() {
	c.CheckService()
}

func (c ClientPeer) CheckService() {
	err := c.Request(c.Name+".Check", "HEAD", "/", nil).
		Do().
		Into(nil)
	statusErr := git_chinawayltd_com_golib_tools_courier_status_error.FromError(err)
	if statusErr.Code == int64(git_chinawayltd_com_golib_tools_courier_status_error.RequestTimeout) {
		panic(fmt.Errorf("service %s have some error %s", c.Name, statusErr))
	}
}

type CreatePeerBody struct {
	PeerID   string `json:"peerID"`
	PeerAddr string `json:"peerAddr"`
}

type CreatePeerRequest struct {
	Body CreatePeerBody `name:"body" in:"body"`
}

func (c ClientPeer) CreatePeer(req CreatePeerRequest, metas ...git_chinawayltd_com_golib_tools_courier.Metadata) (resp *CreatePeerResponse, err error) {
	resp = &CreatePeerResponse{}
	resp.Meta = git_chinawayltd_com_golib_tools_courier.Metadata{}

	err = c.Request(c.Name+".CreatePeer", "POST", "/packager/v0/peers", req, metas...).
		Do().
		BindMeta(resp.Meta).
		Into(&resp.Body)

	return
}

type CreatePeerResponse struct {
	Meta git_chinawayltd_com_golib_tools_courier.Metadata
	Body interface{}
}
