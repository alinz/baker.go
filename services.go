package baker

import (
	"math/rand"
)

type containerValue struct {
	container Container
	endpoint  *Endpoint
}

type Service struct {
	containers []*containerValue
}

func (s *Service) Add(endpoint *Endpoint, container Container) {
	targetID := container.ID()

	value := &containerValue{
		container: container,
		endpoint:  endpoint,
	}

	for i, c := range s.containers {
		if c.container.ID() == targetID {
			s.containers[i] = value
			return
		}
	}

	s.containers = append(s.containers, value)
}

func (s *Service) Remove(container Container) {
	targetID := container.ID()

	for i, c := range s.containers {
		if c.container.ID() == targetID {
			// remove the container and keep the order of the list
			s.containers = append(s.containers[:i], s.containers[i+1:]...)
		}
	}
}

func (s *Service) Random() (Container, *Endpoint) {
	n := len(s.containers)
	if n == 0 {
		return nil, nil
	}

	value := s.containers[rand.Intn(n)]
	return value.container, value.endpoint
}

func (s *Service) Size() int {
	return len(s.containers)
}

func NewService() *Service {
	return &Service{
		containers: make([]*containerValue, 0),
	}
}

type Services struct {
	paths *Paths
}

func (s *Services) Add(endpoint *Endpoint, container Container) {
	key := []rune(endpoint.Path)

	service := s.paths.Get(key)
	if service == nil {
		service = NewService()
		s.paths.Put(key, service)
	}

	service.Add(endpoint, container)
}

func (s *Services) Remove(endpoint *Endpoint, container Container) {
	key := []rune(endpoint.Path)

	service := s.paths.Get(key)
	if service == nil {
		return
	}

	service.Remove(container)

	if len(service.containers) == 0 {
		s.paths.Del(key)
	}
}

func (s *Services) Get(path string) (Container, *Endpoint) {
	key := []rune(path)

	service := s.paths.Get(key)
	if service == nil {
		return nil, nil
	}

	return service.Random()
}

func NewServices() *Services {
	return &Services{
		paths: NewPaths(),
	}
}
