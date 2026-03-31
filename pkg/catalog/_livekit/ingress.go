package livekit

import (
	"context"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/node"
	"github.com/progrium/rig/pkg/resource"
)

type Ingress livekit.IngressInfo

func (r Ingress) Entity() node.E {
	return r
}

func (r Ingress) GetName() string {
	return r.Name
}

func (r Ingress) GetID() string {
	return r.IngressId
}

type ingressProvider struct {
	client *lksdk.IngressClient
}

func (p *ingressProvider) List(ctx context.Context) (resources []resource.Resource[Ingress], err error) {
	lst, err := p.client.ListIngress(ctx, &livekit.ListIngressRequest{})
	if err != nil {
		return nil, err
	}
	for _, i := range lst.Items {
		ii := Ingress(*i)
		resources = append(resources, resource.New(p, ii.GetID(), ii))
	}
	return
}

func (p *ingressProvider) Create(ctx context.Context, param *livekit.CreateIngressRequest) (resource.Resource[Ingress], error) {
	i, err := p.client.CreateIngress(ctx, param)
	if err != nil {
		return resource.Resource[Ingress]{}, err
	}
	ii := Ingress(*i)
	return resource.New(p, ii.GetID(), ii), nil
}

func (p *ingressProvider) Read(ctx context.Context, id resource.ID) (*Ingress, error) {
	return resource.ReadFromList[Ingress](ctx, p, id)
}

func (p *ingressProvider) Delete(ctx context.Context, id resource.ID) error {
	_, err := p.client.DeleteIngress(ctx, &livekit.DeleteIngressRequest{
		IngressId: string(id),
	})
	if err != nil {
		return err
	}
	return nil
}

type IngressList struct {
	Provider *Provider
}

func (l *IngressList) CanAddNode() []string {
	return []string{"Ingress"}
}

func (l *IngressList) AddNode(typ string, parent manifold.Node, oldView string) (bool, error) {
	if typ == "Ingress" {
		resource.NewNode[livekit.CreateIngressRequest, Ingress, IngressList]("New Ingress", parent, oldView, l.Provider.IngressProvider())
		return true, nil
	}
	return false, nil
}

func (l *IngressList) Nodes(com manifold.Node) (nodes node.Nodes) {
	return resource.ListNodes(com, l.Provider.IngressProvider())
}
