package lq

import (
	"sync"
)

const ChanSize = 10

type IWaitGroup interface {
	Add(i int)
	Done()
	Wait()
}

type IListQueue[T any] interface {
	Add(...T)
	Each() chan T
	Close()
	Wait()
}

// NewListQueue returns a queue which ensures exactly once, in order delivery of every past and future event to every
// subscriber.
//
//    		q := listQueue.NewListQueue[int]()
//    		q.Add(1, 2)
//    		ch := q.Each()
//
// q.Each() will return a channel which can be used to iterate over past and future arguments added with Add(...).
// For each item returned from this channel. The queue needs to be closed after all items have been added with Add(...)
// for loops using q.Each() to exit.
//
//    	    go func() {
// 	  	      for v := range ch {
// 	  	      	    // Prints: 		1
// 	  	        	//				2
// 	  	        	fmt.Println(v)
// 	  	        	q.Done()
// 	  	      }
//
// Loops created after Close() is called will still work as expected, but new values can't be added with Add(...).
//
//   		  q.Add(3, 4)
//   		  q.Close()
//   		  for v := range q.Each() {
//   		  	    // Prints: 		1
//   		    	//				2
//   		    	//				3
//   		    	//				4
//   		    	fmt.Println(v)
//   		    	q.Done()
//   		  }
//    	}()
//
//    	// Waits for all events to be delivered to all channels
//    	q.Wait()
//
func NewListQueue[T any](items ...T) *ListQueue[T] {
	return &ListQueue[T]{
		items: items,
		m:     &sync.RWMutex{},
		w:     &sync.WaitGroup{},
	}
}

type ListQueue[T any] struct {
	items     []T
	m         *sync.RWMutex
	w         IWaitGroup
	listeners []chan T
}

// Each returns a channel which will be sent all current and future items passed to Add in the order they were added.
func (s *ListQueue[T]) Each() chan T {
	s.w.Add(1)

	s.m.RLock()
	// Grab the number of items at the moment to ensure we don't double process messages in the case we finish iterating
	// the backlog between the time the backlog updates and items start getting fed to the listeners in s.Add.
	numItems := len(s.items)

	// Create a listener right early, so we don't miss any messages added when we are processing the backlog.
	listener := make(chan T, numItems+ChanSize)
	s.listeners = append(s.listeners, listener)
	s.m.RUnlock()

	go func() {
		for i := 0; i < numItems; i++ {
			listener <- s.items[i]
		}
		s.w.Done()
	}()

	return listener
}

// Add adds all passsed arguments to the list and wakes up any paused goroutines to continue reading from the list.
func (s *ListQueue[T]) Add(items ...T) {
	s.m.Lock()
	s.items = append(s.items, items...)
	s.m.Unlock()

	for _, item := range items {
		for _, l := range s.listeners {
			l <- item
		}
	}
}

func (s *ListQueue[T]) Close() {
	// Wait for backlogs to be copied to listeners.
	s.w.Wait()
	for _, l := range s.listeners {
		close(l)
	}
}
