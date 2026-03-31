package livekit

import (
	"context"
	"fmt"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/node"
	"github.com/progrium/rig/pkg/resource"
)

type Room livekit.Room

func (r Room) Entity() node.E {
	return r
}

func (r Room) GetName() string {
	return r.Name
}

func (r Room) GetID() string {
	return r.Sid
}

type roomProvider struct {
	client *lksdk.RoomServiceClient
}

func (p *roomProvider) List(ctx context.Context) (resources []resource.Resource[Room], err error) {
	lst, err := p.client.ListRooms(ctx, &livekit.ListRoomsRequest{})
	if err != nil {
		return nil, err
	}
	for _, r := range lst.Rooms {
		resources = append(resources, resource.New(p, r.Sid, Room(*r)))
	}
	return
}

func (p *roomProvider) Create(ctx context.Context, room *livekit.CreateRoomRequest) (resource.Resource[Room], error) {
	r, err := p.client.CreateRoom(ctx, room)
	if err != nil {
		return resource.Resource[Room]{}, err
	}
	return resource.New(p, r.Sid, Room(*r)), nil
}

func (p *roomProvider) Read(ctx context.Context, id resource.ID) (*Room, error) {
	return resource.ReadFromList[Room](ctx, p, id)
}

func (p *roomProvider) Delete(ctx context.Context, id resource.ID) error {
	// todo: cache mapping and only do this if not found
	lst, err := p.client.ListRooms(ctx, &livekit.ListRoomsRequest{})
	if err != nil {
		return err
	}
	var name string
	for _, r := range lst.Rooms {
		if r.Sid == string(id) {
			name = r.Name
			break
		}
	}
	if name == "" {
		return fmt.Errorf("unable to find room with id: %s", id)
	}
	_, err = p.client.DeleteRoom(ctx, &livekit.DeleteRoomRequest{
		Room: name,
	})
	return err
}

type RoomList struct {
	Provider *Provider
}

func (l *RoomList) CanAddNode() []string {
	return []string{"Room"}
}

func (l *RoomList) AddNode(typ string, parent manifold.Node, oldView string) (bool, error) {
	if typ == "Room" {
		resource.NewNode[livekit.CreateRoomRequest, Room, RoomList]("New Room", parent, oldView, l.Provider.RoomProvider())
		return true, nil
	}
	return false, nil
}

func (l *RoomList) Nodes(com manifold.Node) (nodes node.Nodes) {
	nodes = resource.ListNodes(com, l.Provider.RoomProvider())
	for _, n := range nodes {
		nn := manifold.FromEntity(n)
		pl := node.Get[*ParticipantList](n)
		if pl == nil {
			room := node.Get[*Room](n)
			nn.AddComponent(ParticipantList{
				Provider: l.Provider,
				Room:     room.Name,
			})
		}
	}
	return
}
