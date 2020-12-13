package sun

import "github.com/sandertv/gophertunnel/minecraft"

type Sun struct {
	Listener *minecraft.Listener
	Players []*Player
	open bool
}

func (s *Sun) main()  {
	for s.open {

	}
}

func (s *Sun) Start()  {
	s.open = true
	go s.main()
}

func (s *Sun) Close()  {
	s.open = false
}