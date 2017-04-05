package mockEtcdClient

import (
	"errors"
	"fmt"
	"time"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type KeysAPIWrapper interface {
	Watcher(key string, opts *client.WatcherOptions) client.Watcher
	Get(ctx context.Context, key string, opts *client.GetOptions) (*client.Response, error)
	Set(ctx context.Context, key, val string, opts *client.SetOptions) (*client.Response, error)
}

type FakeKeysAPI struct {
	ExpectedResponses []*ExpectedResponse
	ExpectedSets      []*ExpectedResponse
	ReceivedSets      []*ExpectedResponse
}

type ExpectedResponse struct {
	Err      error
	Response *client.Response
}

func (e *ExpectedResponse) WillReturnError(err error) *ExpectedResponse {
	e.Err = err
	return e
}

func (e *ExpectedResponse) WillReturnValue(value string) *ExpectedResponse {
	e.Response.Node.Value = value
	return e
}

func (f *FakeKeysAPI) Get(ctx context.Context, key string, opts *client.GetOptions) (resp *client.Response, err error) {
	resp = &client.Response{}
	if len(f.ExpectedResponses) == 0 {
		return nil, errors.New(fmt.Sprintf("Unexpected key Get for %v", key))
	}

	expectedResponse := f.ExpectedResponses[0]
	f.ExpectedResponses = f.ExpectedResponses[1:]

	if expectedResponse.Err != nil {
		return nil, expectedResponse.Err
	}
	if expectedResponse.Response == nil { //!= key {
		return nil, errors.New("bad expectation in mock, Response not found but error was nil")
	}
	if expectedResponse.Response.Node == nil || expectedResponse.Response.Node.Key != key {
		return nil, errors.New(fmt.Sprintf("100: Key not found (%v) [39881395]", key))
	}
	return expectedResponse.Response, nil
}

func (f *FakeKeysAPI) Set(ctx context.Context, key, val string, opts *client.SetOptions) (*client.Response, error) {
	if len(f.ExpectedSets) < 1 {
		return nil, errors.New("no more SET operations expected")
	}
	expected := f.ExpectedSets[0]
	f.ExpectedSets = f.ExpectedSets[1:]
	if expected.Err != nil {
		return nil, expected.Err
	}
	if expected.Response != nil && expected.Response.Node != nil {
		if expected.Response.Node.Key == key && expected.Response.Node.Value == val {
			return expected.Response, nil
		}
		return nil, errors.New(fmt.Sprintf("wrong key/value pair in Set, expected a set for %v/%v, got %v/%v", expected.Response.Node.Key, expected.Response.Node.Value, key, val))
	}
	return nil, errors.New("malformed expected-Set")
}

func (f *FakeKeysAPI) Delete(ctx context.Context, key string, opts *client.DeleteOptions) (*client.Response, error) {
	return nil, errors.New("mock Delete not implemented")
}

func (f *FakeKeysAPI) Create(ctx context.Context, key, value string) (*client.Response, error) {
	return nil, errors.New("mock Create not implemented")
}

func (f *FakeKeysAPI) CreateInOrder(ctx context.Context, dir, value string, opts *client.CreateInOrderOptions) (*client.Response, error) {
	return nil, errors.New("mock CreateInOrder not implemented")
}

func (f *FakeKeysAPI) Update(ctx context.Context, key, value string) (*client.Response, error) {
	return nil, errors.New("mock Update not implemented")
}

func (f *FakeKeysAPI) Watcher(key string, opts *client.WatcherOptions) client.Watcher {
	watcher := &MockWatcher{}
	go func(watcher *MockWatcher) {
		for {
			<-time.After(500 * time.Microsecond)
			for i := 0; i < len(f.ExpectedResponses); i++ {
				if f.ExpectedResponses[i].Response != nil && f.ExpectedResponses[i].Response.Node != nil {
					watcher.Nodes = append(watcher.Nodes, f.ExpectedResponses[i].Response.Node)
				}
			}
		}
	}(watcher)
	return watcher
}

func (f *FakeKeysAPI) ExpectationsFulfilled() error {
	if len(f.ExpectedResponses) != 0 {
		return errors.New("unmet expectations in FakeKeysAPI")
	}
	if len(f.ExpectedSets) != len(f.ReceivedSets) {
		return errors.New("one or more unfulfilled expected sets")
	}
	for i := 0; i < len(f.ExpectedSets); i++ {
		expectedKey := f.ExpectedSets[i].Response.Node.Key
		expectedValue := f.ExpectedSets[i].Response.Node.Value
		receivedKey := f.ReceivedSets[i].Response.Node.Key
		receivedValue := f.ReceivedSets[i].Response.Node.Value
		if expectedKey != receivedKey || expectedValue != receivedValue {
			return errors.New(fmt.Sprintf("expected Sets run out of order, %v, %v, %v, %v", expectedKey, receivedKey, expectedValue, receivedValue))
		}
	}
	return nil
}

func (f *FakeKeysAPI) ExpectGet(key string) *ExpectedResponse {
	node := &client.Node{Key: key}
	resp := &client.Response{Node: node}
	expectedResp := &ExpectedResponse{Response: resp}
	f.ExpectedResponses = append(f.ExpectedResponses, expectedResp)
	return expectedResp
}

func (f *FakeKeysAPI) ExpectSet(key, value string) *ExpectedResponse {
	node := &client.Node{Key: key, Value: value}
	resp := &client.Response{Node: node}
	expectedResp := &ExpectedResponse{Response: resp}
	f.ExpectedSets = append(f.ExpectedSets, expectedResp)
	return expectedResp
}
