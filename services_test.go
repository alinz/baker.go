package baker_test

import (
	"testing"

	"github.com/alinz/baker.go"
)

type DummyContainer struct {
	id   string
	addr string
}

func (d DummyContainer) ID() string {
	return d.id
}

func (d DummyContainer) Addr() string {
	return d.addr
}

func TestService(t *testing.T) {
	dummyContainer1 := DummyContainer{"1", "1.1.1.1"}
	dummyContainer2 := DummyContainer{"2", "1.1.1.2"}
	dummyContainer3 := DummyContainer{"3", "1.1.1.3"}

	dummyEndpoint := &baker.Endpoint{
		Domain: "example.com",
		Path:   "/",
		Ready:  true,
	}

	service := baker.NewService()

	t.Run("it should add 3 containers to service", func(t *testing.T) {
		service.Add(dummyEndpoint, dummyContainer1)
		service.Add(dummyEndpoint, dummyContainer2)
		service.Add(dummyEndpoint, dummyContainer3)

		if service.Size() != 3 {
			t.Errorf("expect 3 but got %d", service.Size())
		}
	})

	t.Run("it should add 3 containers to service, duplicates should be ignored", func(t *testing.T) {
		service.Add(dummyEndpoint, dummyContainer1)
		service.Add(dummyEndpoint, dummyContainer2)
		service.Add(dummyEndpoint, dummyContainer3)
		service.Add(dummyEndpoint, dummyContainer1)
		service.Add(dummyEndpoint, dummyContainer2)
		service.Add(dummyEndpoint, dummyContainer3)

		if service.Size() != 3 {
			t.Errorf("expect 3 but got %d", service.Size())
		}
	})

	t.Run("it should remove one container", func(t *testing.T) {
		service.Remove(dummyContainer1)

		if service.Size() != 2 {
			t.Errorf("expect 2 but got %d", service.Size())
		}
	})

	t.Run("it should randomize getting access to container", func(t *testing.T) {
		read := make(map[string]bool)

		for {
			container, _ := service.Random()
			if len(read) == 2 {
				break
			}

			read[container.ID()] = true
		}
	})

	t.Run("it should return nil if no container available", func(t *testing.T) {
		service.Remove(dummyContainer2)
		service.Remove(dummyContainer3)

		if service.Size() != 0 {
			t.Errorf("expect 0 containers but got %d", service.Size())
		}

		container, _ := service.Random()
		if container != nil {
			t.Errorf("expect no container but got one")
		}
	})
}

func TestServices(t *testing.T) {
	dummyContainer1 := DummyContainer{"1", "1.1.1.1"}
	dummyContainer2 := DummyContainer{"2", "1.1.1.2"}
	dummyContainer3 := DummyContainer{"3", "1.1.1.3"}

	dummyEndpoint1 := &baker.Endpoint{
		Domain: "example.com",
		Path:   "/a/b/c",
		Ready:  true,
	}

	dummyEndpoint2 := &baker.Endpoint{
		Domain: "example.com",
		Path:   "/d/a/v",
		Ready:  true,
	}

	services := baker.NewServices()

	t.Run("it should add 2 paths to services", func(t *testing.T) {
		services.Add(dummyEndpoint1, dummyContainer1)
		services.Add(dummyEndpoint2, dummyContainer2)
		services.Add(dummyEndpoint2, dummyContainer3)

		container, _ := services.Get(dummyEndpoint1.Path)
		if container != dummyContainer1 {
			t.Errorf("expect to get container %s but got %s", dummyContainer1.ID(), container.ID())
		}
	})

	t.Run("it should not return any container because there is no path match", func(t *testing.T) {
		services.Remove(dummyEndpoint1, dummyContainer1)

		container, _ := services.Get(dummyEndpoint1.Path)
		if container != nil {
			t.Errorf("expect to get no containers but got one")
		}
	})

	t.Run("it should get random 2 containers for given path", func(t *testing.T) {
		services.Remove(dummyEndpoint1, dummyContainer1)

		read := make(map[string]bool)

		for {
			if len(read) == 2 {
				break
			}

			container, _ := services.Get(dummyEndpoint2.Path)
			if container == nil {
				t.Errorf("expect to get a container but got none")
			}
			read[container.ID()] = true
		}
	})
}
