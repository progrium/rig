package livekit

import (
	"context"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/progrium/rig/pkg/entity"
	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/resource"
)

type Participant livekit.ParticipantInfo

func (r Participant) Entity() entity.E {
	return r
}

func (r Participant) GetName() string {
	return r.Name
}

func (r Participant) GetID() string {
	return r.Identity
}

type participantProvider struct {
	client *lksdk.RoomServiceClient
	room   string
}

func (p *participantProvider) List(ctx context.Context) (resources []resource.Resource[Participant], err error) {
	lst, err := p.client.ListParticipants(ctx, &livekit.ListParticipantsRequest{
		Room: p.room,
	})
	if err != nil {
		return nil, err
	}
	for _, r := range lst.Participants {
		resources = append(resources, resource.New(p, r.Identity, Participant(*r)))
	}
	return
}

func (p *participantProvider) Read(ctx context.Context, id resource.ID) (*Participant, error) {
	r, err := p.client.GetParticipant(ctx, &livekit.RoomParticipantIdentity{
		Room:     p.room,
		Identity: string(id),
	})
	if err != nil {
		return nil, err
	}
	rr := Participant(*r)
	return &rr, nil
}

func (p *participantProvider) Delete(ctx context.Context, id resource.ID) error {
	_, err := p.client.RemoveParticipant(ctx, &livekit.RoomParticipantIdentity{
		Room:     p.room,
		Identity: string(id),
	})
	return err
}

type ParticipantList struct {
	Provider *Provider
	Room     string
}

func (l *ParticipantList) Nodes(com manifold.Node) (nodes entity.Nodes) {
	return resource.ListNodes(com, l.Provider.ParticipantProvider(l.Room))
}
