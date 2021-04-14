package baker_test

import (
	"testing"

	"github.com/alinz/baker.go"
)

func TestRegistry(t *testing.T) {
	endpoint1 := &baker.Endpoint{"example.com", "/a/b", nil, true}
	endpoint2 := &baker.Endpoint{"api.example.com", "/a/b", nil, true}
	endpoint3 := &baker.Endpoint{"example.com", "/a/b/c", nil, true}

	container1 := &DummyContainer{"1", "1.1.1.1"}
	container2 := &DummyContainer{"2", "1.1.1.2"}
	container3 := &DummyContainer{"3", "1.1.1.3"}

	registry := baker.NewRegistry()

	registry.UpdateContainer(container1, endpoint1)
	registry.UpdateContainer(container2, endpoint2)
	registry.UpdateContainer(container3, endpoint3)

	t.Run("it should return container based on given endpoints", func(t *testing.T) {
		container, _ := registry.FindContainer(endpoint1.Domain, endpoint1.Path)
		if container != container1 {
			t.Fatalf("expect container 1 but got %s", container.ID())
		}

		container, _ = registry.FindContainer(endpoint2.Domain, endpoint2.Path)
		if container != container2 {
			t.Fatalf("expect container 2 but got %s", container.ID())
		}

		container, _ = registry.FindContainer(endpoint3.Domain, endpoint3.Path)
		if container != container3 {
			t.Fatalf("expect container 3 but got %s", container.ID())
		}
	})

	t.Run("it should remove a container from tree", func(t *testing.T) {
		endpoint3.Ready = false

		registry.UpdateContainer(container3, endpoint3)

		container, _ := registry.FindContainer(endpoint3.Domain, endpoint3.Path)
		if container != nil {
			t.Fatalf("expect no container but got one")
		}
	})

	t.Run("it should only add one container despite calling UpdateContainer multiple times", func(t *testing.T) {
		endpoint3.Ready = true

		for i := 0; i < 100; i++ {
			registry.UpdateContainer(container3, endpoint3)
		}

		container, _ := registry.FindContainer(endpoint3.Domain, endpoint3.Path)
		if container == nil {
			t.Fatalf("expect container 3 but got none")
		}
	})
}
