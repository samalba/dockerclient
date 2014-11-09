package dockerclient

import (
	"testing"
)

func TestMock(t *testing.T) {
	mock := NewMockClient()
	mock.On("Version").Return(&Version{Version: "foo"}, nil).Once()

	v, err := mock.Version()
	if err != nil {
		t.Fatal(err)
	}
	if v.Version != "foo" {
		t.Fatal(v)
	}

	mock.Mock.AssertExpectations(t)
}
