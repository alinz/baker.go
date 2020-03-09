package trie_test

import (
	"testing"

	"github.com/alinz/baker.go/internal/store"
	"github.com/alinz/baker.go/internal/store/trie"
)

type keyValue struct {
	key   string
	value string
}

func newKeyValue(key string, value string) keyValue {
	return keyValue{
		key:   key,
		value: value,
	}
}

func compareValues(val1, val2 interface{}) bool {
	str1 := (val1).(string)
	str2 := (val2).(string)

	return str1 == str2
}

func TestPutGetTrie(t *testing.T) {

	testCases := []struct {
		keyValues     []keyValue
		givenKey      string
		expectedValue string
		expectedError error
	}{
		{
			keyValues: []keyValue{
				newKeyValue("/api*", "api-node"),
			},
			givenKey:      "/api/",
			expectedValue: "api-node",
		},
		{
			keyValues: []keyValue{
				newKeyValue("/a*", "api-node1"),
				newKeyValue("/ap*", "api-node2"),
			},
			givenKey:      "/ap/",
			expectedValue: "api-node2",
		},
		{
			keyValues: []keyValue{
				newKeyValue("/ap*", "api-node2"),
				newKeyValue("/a*", "api-node1"),
			},
			givenKey:      "/ap/",
			expectedValue: "api-node2",
		},
		{
			keyValues: []keyValue{
				newKeyValue("/ap*", "api-node2"),
				newKeyValue("/a*", "api-node1"),
			},
			givenKey:      "/ap",
			expectedValue: "api-node2",
		},
	}

	for i, testCase := range testCases {
		root := trie.New(false)

		for _, kv := range testCase.keyValues {
			err := root.Put([]rune(kv.key), kv.value)
			if err != nil {
				t.Fatalf("in testcase %d, failed to insert %s", i+1, kv.key)
			}
		}

		value, err := root.Get([]rune(testCase.givenKey))
		if err != testCase.expectedError {
			t.Fatalf("in testcase %d, expected %s error but got %s", i+1, testCase.expectedError, err)
		}

		if !compareValues(value, testCase.expectedValue) {
			t.Fatalf("in testcase %d, expected %v but got %v", i+1, testCase.expectedValue, value)
		}
	}
}

type action struct {
	kind  string
	key   string
	value string
}

func TestPutDelGet(t *testing.T) {
	testCases := []struct {
		actions       []action
		key           string
		expectedValue string
		expectedError error
	}{
		{
			actions: []action{
				action{
					key:   "/api*",
					value: "api-node1",
					kind:  "put",
				},
				action{
					key:   "/api/node*",
					value: "api-node2",
					kind:  "put",
				},
			},
			key:           "/api/hello",
			expectedValue: "api-node1",
		},
		{
			actions: []action{
				action{
					key:   "/api*",
					value: "api-node1",
					kind:  "put",
				},
				action{
					key:   "/api/node*",
					value: "api-node2",
					kind:  "put",
				},
			},
			key:           "/api/node/hello",
			expectedValue: "api-node2",
		},
		{
			actions: []action{
				action{
					key:   "/api*",
					value: "api-node1",
					kind:  "put",
				},
				action{
					key:   "/api/node*",
					value: "api-node2",
					kind:  "put",
				},
			},
			key:           "/api/node/hello",
			expectedValue: "api-node2",
		},
		{
			actions: []action{
				action{
					key:   "/api*",
					value: "api-node1",
					kind:  "put",
				},
				action{
					key:   "/api/node*",
					value: "api-node2",
					kind:  "put",
				},
				action{
					key:   "/api/node*",
					value: "api-node2",
					kind:  "del",
				},
			},
			key:           "/api/node/hello",
			expectedValue: "api-node2",
			expectedError: store.ErrItemNotFound,
		},
	}

	put := func(t *testing.T, i int, root *trie.Node, key string, value string) {
		err := root.Put([]rune(key), value)
		if err != nil {
			t.Fatalf("in testcase %d, failed to put %s", i+1, key)
		}
	}

	del := func(t *testing.T, i int, root *trie.Node, key string) {
		err := root.Del([]rune(key))
		if err != nil {
			t.Fatalf("in testcase %d, failed to delete %s", i+1, key)
		}
	}

	for i, testCase := range testCases {
		root := trie.New(false)

		for _, action := range testCase.actions {
			switch action.kind {
			case "put":
				put(t, i, root, action.key, action.value)
			case "del":
				del(t, i, root, action.key)
			}
		}

		result, err := root.Get([]rune(testCase.key))
		if err != nil {
			if err != testCase.expectedError {
				t.Fatalf("in testcase %d, expected %s error but got %s", i, testCase.expectedError, err)
			}
		} else {
			if !compareValues(result, testCase.expectedValue) {
				t.Fatalf("in testcase %d, expected %v value but got %v", i, testCase.expectedValue, result)
			}
		}
	}
}
