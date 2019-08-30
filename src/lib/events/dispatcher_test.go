package events

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddAndRemoveEventListener(t *testing.T) {
	d := NewDispatcher(context.Background()).(*dispatcher)
	l := NewEventListener(func(event *Event) {})
	d.AddEventListener("test", l)
	d.AddEventListener("test2", NewEventListener(func(event *Event) {}))
	ls, ok := d.saver["test"]
	assert.True(t, ok)
	assert.Equal(t, l, ls.Front().Value)
	d.RemoveEventListener("test", l)
	_, ok = d.saver["test"]
	assert.False(t, ok)
	d.RemoveAllEventListener("test2")
	assert.Empty(t, d.saver)
}

func TestDispatchEvent(t *testing.T) {
	l := make([]int, 0)
	d := NewDispatcher(context.Background()).(*dispatcher)
	d.AddEventListener("test", NewEventListener(func(event *Event) {
		l = append(l, 0)
	}))
	d.AddEventListener("test", NewEventListener(func(event *Event) {
		l = append(l, 1)
	}))
	d.AddEventListener("test", NewEventListener(func(event *Event) {
		l = append(l, 2)
	}))
	d.AddEventListener("test", NewEventListener(func(event *Event) {
		l = append(l, 3)
	}))
	d.DispatchEvent(NewEvent("test", nil))
	time.Sleep(time.Second)
	assert.Equal(t, []int{0, 1, 2, 3}, l)
}
