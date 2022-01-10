package lq

import (
	"github.com/stretchr/testify/mock"
	"reflect"
	"sync"
	"testing"
)

type MockWaitGroup struct{ mock.Mock }

func (m *MockWaitGroup) Add(i int) { m.Called(i) }
func (m *MockWaitGroup) Done()     { m.Called() }
func (m *MockWaitGroup) Wait()     { m.Called() }

func TestListQueue_Add(t *testing.T) {
	type fields struct {
		work        IWaitGroup
		items       []string
		c           *sync.Cond
		notDone     bool
		subscribers int
	}
	type args struct {
		items            []string
		expectedWork     int
		expectedNumItems int
	}

	tests := []struct {
		name string
		args args
		q    *ListQueue[string]
	}{
		{
			name: "does nothing with no arguments",
			q: &ListQueue[string]{
				work:    &MockWaitGroup{},
				items:   make([]string, 0, 1000),
				c:       sync.NewCond(&sync.Mutex{}),
				notDone: true,
			},
			args: args{
				items:            []string{},
				expectedWork:     0,
				expectedNumItems: 0,
			},
		},
		{
			name: "updates items and adds no expectedWork when there are no subscribers",
			q: &ListQueue[string]{
				work:    &MockWaitGroup{},
				items:   make([]string, 0, 1000),
				c:       sync.NewCond(&sync.Mutex{}),
				notDone: true,
			},
			args: args{
				items:            []string{"a", "b"},
				expectedWork:     0,
				expectedNumItems: 2,
			},
		},
		{
			name: "when there is one subscriber and 3 args add 3 items of expectedWork",
			q: &ListQueue[string]{
				work:        &MockWaitGroup{},
				items:       make([]string, 0, 1000),
				c:           sync.NewCond(&sync.Mutex{}),
				notDone:     true,
				subscribers: 1,
			},
			args: args{
				items:            []string{"a", "b", "c"},
				expectedWork:     3,
				expectedNumItems: 3,
			},
		},
		{
			name: "when called with existing items and 2 subscribers, 2 arguments add 4 items of expectedWork",
			q: &ListQueue[string]{
				work:        &MockWaitGroup{},
				items:       []string{"a", "b"},
				c:           sync.NewCond(&sync.Mutex{}),
				notDone:     true,
				subscribers: 2,
			},
			args: args{
				items:            []string{"c", "d"},
				expectedWork:     4,
				expectedNumItems: 4,
			},
		},
		{
			name: "when called with 2 args and 2 existing items, items becomes 4",
			q: &ListQueue[string]{
				work:        &MockWaitGroup{},
				items:       []string{"a", "b"},
				c:           sync.NewCond(&sync.Mutex{}),
				notDone:     true,
				subscribers: 0,
			},
			args: args{
				items:            []string{"c", "d"},
				expectedWork:     0,
				expectedNumItems: 4,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origItems := make([]string, len(tt.q.items))
			copy(origItems, tt.q.items)
			tt.q.work.(*MockWaitGroup).On("Add", tt.args.expectedWork)
			tt.q.Add(tt.args.items...)
			if tt.args.expectedNumItems != len(tt.q.items) {
				t.Errorf("len(tt.args.expectedNumItems) != len(tt.q.items): %d != %d", tt.args.expectedNumItems, len(tt.q.items))
			}
			for i, v := range append(origItems, tt.args.items...) {
				//off := len(tt.q.items) - (len(tt.args.items) - i) // Only check added items
				if v != tt.q.items[i] {
					t.Errorf("item tt.args.items[%d] != tt.q.items[%d]: %s != %s", i, i, v, tt.q.items[i])
				}
			}
			tt.q.work.(*MockWaitGroup).AssertNumberOfCalls(t, "Add", 1)
		})
	}
}

// TODO: Finish tests
func TestListQueue_Each(t *testing.T) {
	type fields struct {
		work        *sync.WaitGroup
		items       []string
		c           *sync.Cond
		notDone     bool
		subscribers int
	}
	tests := []struct {
		name   string
		fields fields
		want   chan string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ListQueue[string]{
				work:        tt.fields.work,
				items:       tt.fields.items,
				c:           tt.fields.c,
				notDone:     tt.fields.notDone,
				subscribers: tt.fields.subscribers,
			}
			if got := s.Each(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Each() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListQueue_Wait(t *testing.T) {
	type fields struct {
		work        *sync.WaitGroup
		items       []string
		c           *sync.Cond
		notDone     bool
		subscribers int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ListQueue[string]{
				work:        tt.fields.work,
				items:       tt.fields.items,
				c:           tt.fields.c,
				notDone:     tt.fields.notDone,
				subscribers: tt.fields.subscribers,
			}
			s.Wait()
		})
	}
}

func TestListQueue_next(t *testing.T) {
	type fields struct {
		work        *sync.WaitGroup
		items       []string
		c           *sync.Cond
		notDone     bool
		subscribers int
	}
	type args struct {
		i int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ListQueue[string]{
				work:        tt.fields.work,
				items:       tt.fields.items,
				c:           tt.fields.c,
				notDone:     tt.fields.notDone,
				subscribers: tt.fields.subscribers,
			}
			if got := s.next(tt.args.i); got != tt.want {
				t.Errorf("next() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListQueue_sendFrom(t *testing.T) {
	type fields struct {
		work        *sync.WaitGroup
		items       []string
		c           *sync.Cond
		notDone     bool
		subscribers int
	}
	type args struct {
		start int
		ch    chan string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ListQueue[string]{
				work:        tt.fields.work,
				items:       tt.fields.items,
				c:           tt.fields.c,
				notDone:     tt.fields.notDone,
				subscribers: tt.fields.subscribers,
			}
			if got := s.sendFrom(tt.args.start, tt.args.ch); got != tt.want {
				t.Errorf("sendFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewListQueue(t *testing.T) {
	tests := []struct {
		name string
		want *ListQueue[string]
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewListQueue[string](); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewListQueue() = %v, want %v", got, tt.want)
			}
		})
	}
}
