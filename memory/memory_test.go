package cache

import (
	"net/http"
	"testing"
	"time"

	parent "github.com/vodolaz095/gin-cache"
)

var testMemoryStore *MemoryCache

func TestNew(t *testing.T) {
	testMemoryStore = New(time.Second)
	if testMemoryStore.expirationInterval != time.Second {
		t.Error("wrong expiration duration")
	}
	if len(testMemoryStore.items) != 0 {
		t.Error("items present?")
	}
}

func TestMemoryCache_Save(t *testing.T) {
	err := testMemoryStore.Save("a", parent.Data{
		Body:        []byte("this is body of a key"),
		Status:      http.StatusTeapot,
		ContentType: "text/plain",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Second),
	})
	if err != nil {
		t.Error(err)
	}
	if len(testMemoryStore.items) != 1 {
		t.Error("key length is wrong?")
	}
	if testMemoryStore.items["a"].Key != "a" {
		t.Error("wrongly saved?")
	}
}

func TestMemoryCache_Get(t *testing.T) {
	var hit parent.Data
	var err error
	var found bool
	_, found, err = testMemoryStore.Get("key not found")
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Error("key is found?")
	}
	hit, found, err = testMemoryStore.Get("a")
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

func TestMemoryCache_Delete(t *testing.T) {
	err := testMemoryStore.Save("b", parent.Data{
		Body:        []byte("this is body of a key"),
		Status:      http.StatusTeapot,
		ContentType: "text/plain",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Second),
	})
	if err != nil {
		t.Error(err)
	}
	if len(testMemoryStore.items) != 2 {
		t.Error("key length wrong?")
	}
	if testMemoryStore.items["a"].Key != "a" {
		t.Error("wrongly saved?")
	}
	if testMemoryStore.items["b"].Key != "b" {
		t.Error("wrongly saved?")
	}
	err = testMemoryStore.Delete("b")
	if err != nil {
		t.Error(err)
	}
	if len(testMemoryStore.items) != 1 {
		t.Error("key length wrong?")
	}
	if testMemoryStore.items["a"].Key != "a" {
		t.Error("wrongly saved?")
	}
	_, found, err := testMemoryStore.Get("b")
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
	if len(testMemoryStore.items) != 0 {
		t.Error("key length wrong?")
	}
	_, found, err := testMemoryStore.Get("a")
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Error("deleted key is found?")
	}
}
