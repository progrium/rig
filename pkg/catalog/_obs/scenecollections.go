package obs

import (
	"context"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/config"
	"github.com/progrium/rig/pkg/resource"
)

type SceneCollection struct {
	Name string
}

type sceneCollectionProvider struct {
	client *goobs.Client
}

func (res *sceneCollectionProvider) List(ctx context.Context) (resources []resource.Resource[SceneCollection], err error) {
	lst, err := res.client.Config.GetSceneCollectionList()
	if err != nil {
		return nil, err
	}
	for _, sc := range lst.SceneCollections {
		resources = append(resources, resource.New(res, sc, SceneCollection{Name: sc}))
	}
	return
}

func (res *sceneCollectionProvider) Create(ctx context.Context, params *config.CreateSceneCollectionParams) (resource.Resource[SceneCollection], error) {
	_, err := res.client.Config.CreateSceneCollection(params)
	if err != nil {
		return resource.Resource[SceneCollection]{}, err
	}
	return resource.New(res, *params.SceneCollectionName, SceneCollection{Name: *params.SceneCollectionName}), nil
}

func (res *sceneCollectionProvider) Read(ctx context.Context, id resource.ID) (*SceneCollection, error) {
	return resource.ReadFromList[SceneCollection](ctx, res, id)
}

// func (res *sceneCollectionProvider) Delete(ctx context.Context, id resource.ID) error {

// }
