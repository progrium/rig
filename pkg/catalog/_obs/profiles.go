package obs

import (
	"context"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/config"
	"github.com/progrium/rig/pkg/resource"
)

type Profile struct {
	Name string
}

type profileProvider struct {
	client *goobs.Client
}

func (res *profileProvider) List(ctx context.Context) (resources []resource.Resource[Profile], err error) {
	lst, err := res.client.Config.GetProfileList()
	if err != nil {
		return nil, err
	}
	for _, p := range lst.Profiles {
		resources = append(resources, resource.New(res, p, Profile{Name: p}))
	}
	return
}

func (res *profileProvider) Create(ctx context.Context, params *config.CreateProfileParams) (resource.Resource[Profile], error) {
	_, err := res.client.Config.CreateProfile(params)
	if err != nil {
		return resource.Resource[Profile]{}, err
	}
	return resource.New(res, *params.ProfileName, Profile{Name: *params.ProfileName}), nil
}

func (res *profileProvider) Read(ctx context.Context, id resource.ID) (*Profile, error) {
	return resource.ReadFromList[Profile](ctx, res, id)
}

func (res *profileProvider) Delete(ctx context.Context, id resource.ID) error {
	name := string(id)
	_, err := res.client.Config.RemoveProfile(&config.RemoveProfileParams{
		ProfileName: &name,
	})
	if err != nil {
		return err
	}
	return nil
}
