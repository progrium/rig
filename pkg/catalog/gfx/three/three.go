package three

import "github.com/progrium/rig/pkg/catalog/gfx"

type Scene struct{}

type Camera struct{}

type Transform struct {
	Position gfx.Vector3
	Rotation gfx.Vector3
	Scale    gfx.Vector3
}

type PlaneGeometry struct {
	Width, Height float64
}

type SphereGeometry struct {
	Radius float64
}

type TeapotGeometry struct{}

type Material struct {
	Color string
}

type MeshRenderer struct {
}

type DirectionalLight struct {
	Color     string
	Intensity float64
}

type PointLight struct {
	Color     string
	Intensity float64
	Distance  float64
}
