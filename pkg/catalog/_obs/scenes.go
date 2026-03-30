package obs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/progrium/rig/pkg/entity"
	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/node"
	"github.com/progrium/rig/pkg/resource"
	"github.com/qri-io/jsonpointer"
)

func selectValue(msg json.RawMessage, path string) any {
	b, err := msg.MarshalJSON()
	if err != nil {
		panic(err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		panic(err)
	}
	ptr, err := jsonpointer.Parse(path)
	if err != nil {
		panic(err)
	}
	got, err := ptr.Eval(out)
	if err != nil {
		panic(err)
	}
	return got
}

type Scene struct {
	UUID    string
	Name    string
	Index   int
	Program bool
	Preview bool
}

func (r Scene) GetName() string {
	return r.Name
}

func (r Scene) GetID() string {
	return r.UUID
}

type sceneProvider struct {
	client *goobs.Client
}

func (p *sceneProvider) List(ctx context.Context) (resources []resource.Resource[Scene], err error) {
	lst, err := p.client.Scenes.GetSceneList()
	if err != nil {
		return nil, err
	}
	for idx, scene := range lst.Scenes {
		uuid := selectValue(lst.GetRaw(), fmt.Sprintf("/scenes/%d/sceneUuid", idx)).(string)
		s := Scene{
			UUID:    uuid,
			Name:    scene.SceneName,
			Index:   scene.SceneIndex,
			Program: lst.CurrentProgramSceneUuid == uuid,
			Preview: lst.CurrentPreviewSceneUuid == uuid,
		}
		resources = append(resources, resource.New(p, uuid, s))
	}
	return
}

func (p *sceneProvider) Create(ctx context.Context, params *scenes.CreateSceneParams) (resource.Resource[Scene], error) {
	resp, err := p.client.Scenes.CreateScene(params)
	if err != nil {
		return resource.Resource[Scene]{}, err
	}
	r, err := p.Read(ctx, resource.ID(resp.SceneUuid))
	if err != nil {
		return resource.Resource[Scene]{}, err
	}
	return resource.New(p, resp.SceneUuid, *r), nil
}

func (p *sceneProvider) Read(ctx context.Context, id resource.ID) (*Scene, error) {
	return resource.ReadFromList[Scene](ctx, p, id)
}

func (p *sceneProvider) Delete(ctx context.Context, id resource.ID) error {
	uuid := string(id)
	_, err := p.client.Scenes.RemoveScene(&scenes.RemoveSceneParams{
		SceneUuid: &uuid,
	})
	return err
}

type SceneList struct {
	provider *sceneProvider

	Client *Client // todo: find better way to setup SceneItemList
}

func (l *SceneList) CanAddNode() []string {
	return []string{"Scene"}
}

func (l *SceneList) AddNode(typ string, parent manifold.Node, oldView string) (bool, error) {
	if typ == "Scene" {
		resource.NewNode[scenes.CreateSceneParams, Scene, SceneList]("New Scene", parent, oldView, l.provider)
		return true, nil
	}
	return false, nil
}

func (l *SceneList) Nodes(com manifold.Node) (nodes entity.Nodes) {
	nodes = resource.ListNodes[Scene](com, l.provider)
	for _, n := range nodes {
		nn := manifold.FromEntity(n)
		scene := node.Get[*Scene](n)
		nn.SetAttr("desc", "")
		if scene.Preview {
			nn.SetAttr("desc", "[preview]")
		}
		if scene.Program {
			nn.SetAttr("desc", "[program]")
		}
		pl := node.Get[*SceneItemList](n)
		if pl == nil {
			nn.AddComponent(SceneItemList{Client: l.Client, SceneName: scene.Name})
			nn.SetAttr("view", "obs.SceneItemList")
		}
	}
	return
}
