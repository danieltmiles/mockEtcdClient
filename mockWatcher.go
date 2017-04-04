package mockEtcdClient

import (
	"errors"
	"fmt"
	"time"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type MockWatcher struct {
	Nodes []*client.Node
}

func (f *MockWatcher) ExpectResponse(node *client.Node) {
	f.Nodes = append(f.Nodes, node)
}

func (f *MockWatcher) ExpectationsWereFulfilled() error {
	if len(f.Nodes) != 0 {
		return errors.New(fmt.Sprintf("Unmet expectations in Watcher: %+v", f.Nodes))
	}
	return nil
}

func (f *MockWatcher) Next(ctx context.Context) (*client.Response, error) {
	for len(f.Nodes) < 1 {
		<-time.After(time.Millisecond)
	}
	resp := client.Response{
		Node: f.Nodes[0],
	}
	f.Nodes = f.Nodes[1:]
	return &resp, nil
}
