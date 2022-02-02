package lq

import (
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
)

type MockWaitGroup struct{ mock.Mock }

func (m *MockWaitGroup) Add(i int) { m.Called(i) }
func (m *MockWaitGroup) Done()     { m.Called() }
func (m *MockWaitGroup) Wait()     { m.Called() }

func TestListQueue_Add(t *testing.T) {
	type args struct {
		items            []int
		expectedNumItems int
	}

	tests := []struct {
		name string
		args args
		q    *ListQueue[int]
	}{
		{
			name: "does nothing with no arguments",
			q:    NewListQueue[int](),
			args: args{
				items:            []int{},
				expectedNumItems: 0,
			},
		},
		{
			name: "adds new items to empty backlog",
			q:    NewListQueue[int](),
			args: args{
				items:            []int{1, 2},
				expectedNumItems: 2,
			},
		},
		{
			name: "backlog stays the same between empty Add calls",
			q: &ListQueue[int]{
				items: []int{1, 2},
				m:     &sync.RWMutex{},
				w:     &MockWaitGroup{},
			},
			args: args{
				items:            []int{},
				expectedNumItems: 2,
			},
		},
		{
			name: "adds new items to non-empty backlog",
			q: &ListQueue[int]{
				items: []int{1, 2, 3},
				m:     &sync.RWMutex{},
				w:     &MockWaitGroup{},
			},
			args: args{
				items:            []int{4, 5, 6},
				expectedNumItems: 6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origItems := make([]int, len(tt.q.items))
			copy(origItems, tt.q.items)

			tt.q.Add(tt.args.items...)
			if tt.args.expectedNumItems != len(tt.q.items) {
				t.Errorf("len(tt.args.expectedNumItems) != len(tt.q.items): %d != %d", tt.args.expectedNumItems, len(tt.q.items))
			}
			for i, v := range append(origItems, tt.args.items...) {
				if v != tt.q.items[i] {
					t.Errorf("item tt.args.items[%d] != tt.q.items[%d]: %d != %d", i, i, v, tt.q.items[i])
				}
			}
		})
	}
}

// TODO: Finish tests
func TestNewListQueue(t *testing.T) {
	tests := []struct {
		name    string
		q       *ListQueue[int]
		addArgs []int
		want    []int
	}{
		{
			name:    "Each() chan returns nothing when no items are added",
			q:       NewListQueue[int](),
			addArgs: []int{},
			want:    []int{},
		},
		{
			name:    "Each() chan returns initialized items when none are added",
			q:       NewListQueue[int](1, 2, 3, 4, 5),
			addArgs: []int{},
			want:    []int{1, 2, 3, 4, 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := tt.q.Each()
			tt.q.Add(tt.addArgs...)

			// Should match the number of times Each() is called.
			if len(tt.q.listeners) != 1 {
				t.Errorf("len(listeners) = %d, expected 1", len(tt.q.listeners))
			}
			go checkQueue(t, ch, tt)

			ch2 := tt.q.Each()
			tt.q.Add(tt.addArgs...)
			tt.q.Close()

			if len(tt.q.listeners) != 2 {
				t.Errorf("len(listeners) = %d, expected 1", len(tt.q.listeners))
			}

			checkQueue(t, ch2, tt)
		})
	}
}

func checkQueue(t *testing.T, ch chan int, tt struct {
	name    string
	q       *ListQueue[int]
	addArgs []int
	want    []int
}) {
	var i int
	for item := range ch {
		if item != tt.want[i] {
			t.Errorf("Each() item number %d = %d, expected %d", i, item, tt.want[i])
		}
		i++
	}
	if i != len(tt.want) {
		t.Errorf("Each() chan produced %d results, expected %d", i, len(tt.want))
	}
}
