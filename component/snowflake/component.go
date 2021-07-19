package cherrySnowflake

import (
	cherryFacade "github.com/cherry-game/cherry/facade"
)

type Component struct {
	cherryFacade.Component
}

func NewComponent() *Component {
	return &Component{}
}

func (s *Component) Name() string {
	return "snow_flake_component"
}

func (s *Component) Init() {
	id := s.App().Settings().Get("snow_flake_unique_id")
	if id.LastError() != nil {
		panic("`snow_flake_unique_id` property not found.")
	}

	SetDefaultNode(id.ToInt64())
}
