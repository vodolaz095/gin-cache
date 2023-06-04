package rcache

import (
	"context"
	"net/http"
	"testing"
	"time"

	parent "github.com/vodolaz095/gin-cache"
)

var testMemoryStore *Cache
var testContext context.Context

func TestNew(t *testing.T) {
	var err error
	testMemoryStore, err = New(DefaultConnectionString, "HolyMeat")
	if err != nil {
		t.Errorf("%s : while dialing redis", err)
	}
	testContext = context.TODO()
}

func TestCache_Save(t *testing.T) {
	err := testMemoryStore.Save(testContext, "a", parent.Data{
		Body:        []byte("this is body of a key"),
		Status:      http.StatusTeapot,
		ContentType: "text/plain",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Second),
	})
	if err != nil {
		t.Error(err)
	}
}

func TestCache_Get(t *testing.T) {
	var hit parent.Data
	var err error
	var found bool
	_, found, err = testMemoryStore.Get(testContext, "key not found")
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Error("key is found?")
	}
	hit, found, err = testMemoryStore.Get(testContext, "a")
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Error("key is not found?")
	}
	if hit.Key != "a" {
		t.Error("wrongly saved?")
	}
	if string(hit.Body) != "this is body of a key" {
		t.Error("wrongly saved?")
	}
}

func TestCache_Delete(t *testing.T) {
	var err error
	var found bool
	err = testMemoryStore.Save(testContext, "b", parent.Data{
		Body:        []byte("this is body of a key"),
		Status:      http.StatusTeapot,
		ContentType: "text/plain",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Second),
	})
	if err != nil {
		t.Error(err)
	}
	_, found, err = testMemoryStore.Get(testContext, "b")
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Error("key not found")
	}

	err = testMemoryStore.Delete(testContext, "b")
	if err != nil {
		t.Error(err)
	}
	_, found, err = testMemoryStore.Get(testContext, "b")
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Error("deleted key is found?")
	}
}

func TestExpires(t *testing.T) {
	time.Sleep(time.Second)
	time.Sleep(time.Second)
	_, found, err := testMemoryStore.Get(testContext, "a")
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Error("deleted key is found?")
	}
}
