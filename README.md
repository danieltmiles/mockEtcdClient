# mockEtcdClient

## Mock library to simulate interactions with etcd as if by the github.com/coreos/etcd/client library.

### Starting with a faked KeysAPI, tests look like this (from mock_test.go):
```go
package mockEtcdClient

import (
	"errors"
	"fmt"
	"testing"

	"github.com/coreos/etcd/client"
	. "github.com/onsi/gomega"
	"golang.org/x/net/context"
)

func TestExpectGet(t *testing.T) {
	RegisterTestingT(t)
	var kapi client.KeysAPI
	mock := FakeKeysAPI{}
	kapi = &mock
	mock.ExpectGet("/test/key").WillReturnValue("test value")

	resp, err := kapi.Get(context.Background(), "/test/key", nil)

	Expect(err).NotTo(HaveOccurred())
	Expect(resp).NotTo(BeNil())
	Expect(resp.Node).NotTo(BeNil())
	Expect(resp.Node.Key).To(Equal("/test/key"))
	Expect(resp.Node.Value).To(Equal("test value"))
	err = mock.ExpectationsFulfilled()
	Expect(err).NotTo(HaveOccurred())
}

func TestExpectErr(t *testing.T) {
	RegisterTestingT(t)
	var kapi client.KeysAPI
	mock := FakeKeysAPI{}
	kapi = &mock
	key := "/test/key"
	mock.ExpectGet(key).WillReturnError(errors.New(fmt.Sprintf("100: Key not found (%v) [39881395]", key)))

	resp, err := kapi.Get(context.Background(), "/test/key", nil)
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal(fmt.Sprintf("100: Key not found (%v) [39881395]", key)))
	Expect(resp).To(BeNil())
}

func TestExpectSet(t *testing.T) {
	RegisterTestingT(t)
	var kapi client.KeysAPI
	mock := FakeKeysAPI{}
	kapi = &mock

	mock.ExpectSet("/some/key", "some value")
	resp, err := kapi.Set(context.Background(), "/some/key", "some value", nil)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.Node).NotTo(BeNil())
	Expect(resp.Node.Key).To(Equal("/some/key"))
	Expect(resp.Node.Value).To(Equal("some value"))
	err = mock.ExpectationsFulfilled()
	Expect(err).NotTo(HaveOccurred())
}

func TestExpectSetError(t *testing.T) {
	RegisterTestingT(t)
	var kapi client.KeysAPI
	mock := FakeKeysAPI{}
	kapi = &mock

	mock.ExpectSet("/some/key", "some value").WillReturnError(errors.New("this is a testing error"))
	resp, err := kapi.Set(context.Background(), "/some/other/key", "some value", nil)
	Expect(err).To(HaveOccurred())
	Expect(resp).To(BeNil())
	err = mock.ExpectationsFulfilled()
	Expect(err).NotTo(HaveOccurred())
}

func TestExpectSetNotRecvd(t *testing.T) {
	RegisterTestingT(t)
	mock := FakeKeysAPI{}

	mock.ExpectSet("/some/key", "some value")
	err := mock.ExpectationsFulfilled()
	Expect(err).To(HaveOccurred())
}
```
