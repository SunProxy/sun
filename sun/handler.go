package sun

import (
	"github.com/sunproxy/sun/sun/event"
	"github.com/sunproxy/sun/sun/planet"
	"github.com/sunproxy/sun/sun/ray"
)

type Handler interface {
	HandleRayJoin(ctx *event.Context, ray *ray.Ray)
	HandlePlanetConnect(ctx *event.Context, planet *planet.Planet)
}

type NopHandler struct{}

func (n NopHandler) HandleRayJoin(*event.Context, *ray.Ray) {}

func (n NopHandler) HandlePlanetConnect(*event.Context, *planet.Planet) {}

// Compile time check to make sure NopHandler implements Handler.
var _ Handler = (*NopHandler)(nil)
