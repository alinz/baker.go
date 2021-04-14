package baker_test

import (
	"testing"

	"github.com/alinz/baker.go"
)

func TestPaths(t *testing.T) {
	container1 := DummyContainer{"1", "1.1.1.1"}
	container2 := DummyContainer{"2", "1.1.1.2"}

	endpoint1 := &baker.Endpoint{
		Domain: "example.com",
		Path:   "/a/b/c",
		Ready:  true,
	}

	endpoint2 := &baker.Endpoint{
		Domain: "example.com",
		Path:   "/a/b",
		Ready:  true,
	}

	service1 := baker.NewService()
	service2 := baker.NewService()

	service1.Add(endpoint1, container1)
	service2.Add(endpoint2, container2)

	paths := baker.NewPaths()

	paths.Put([]rune(endpoint1.Path), service1)
	paths.Put([]rune(endpoint2.Path), service2)

	t.Run("it should find proper container based on given path", func(t *testing.T) {
		service := paths.Get([]rune(endpoint1.Path))

		container, _ := service.Random()
		if container != container1 {
			t.Errorf("expect container 1 but got %s", container.ID())
		}

		service = paths.Get([]rune(endpoint2.Path))
		container, _ = service.Random()
		if container != container2 {
			t.Errorf("expect container 2 but got %s", container.ID())
		}
	})

	t.Run("it should find proper container based on given path", func(t *testing.T) {
		paths.Del([]rune(endpoint1.Path))

		service := paths.Get([]rune(endpoint2.Path))

		container, _ := service.Random()
		if container != container2 {
			t.Errorf("expect container 2 but got %s", container.ID())
		}

		paths.Del([]rune(endpoint2.Path))
		service = paths.Get([]rune(endpoint2.Path))
		if service != nil {
			t.Errorf("expect no service but got one")
		}
	})
}
