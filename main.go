package lq

import (
	"sync"
)

const ChanSize = 1000

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
	wg := &sync.WaitGroup{}
	wg.Add(1)

	return &ListQueue[T]{
		items: items,
		keys:  map[string]T{},
		mKeys: &sync.RWMutex{},
		m:     &sync.RWMutex{},
		Wg:    wg,
	}
}

type ListQueue[T any] struct {
	items     []T
	keys      map[string]T
	mKeys     *sync.RWMutex
	m         *sync.RWMutex
	Wg        IWaitGroup
	listeners []chan T
}

// Each returns a channel which will be sent all current and future items passed to Add in the order they were added.
func (lq *ListQueue[T]) Each() chan T {
	lq.m.RLock()
	// Grab the number of items at the moment to ensure we don't double process messages in the case we finish iterating
	// the backlog between the time the backlog updates and items start getting fed to the listeners in s.Add.
	numItems := len(lq.items)
	lq.Wg.Add(numItems)

	// Create a listener right early, so we don't miss any messages added when we are processing the backlog.
	writer := make(chan T, numItems+ChanSize)
	lq.listeners = append(lq.listeners, writer)
	lq.m.RUnlock()

	reader := make(chan T, 0)

	go func() {
		for i := range writer {
			reader <- i
			lq.Wg.Done()
		}
	}()
	go func() {
		for i := 0; i < numItems; i++ {
			writer <- lq.items[i]
		}
	}()

	return reader
}

// Add adds all passsed arguments to the list and wakes up any paused goroutines to continue reading from the list.
func (lq *ListQueue[T]) Add(items ...T) {
	lq.m.Lock()
	work := len(lq.listeners) * len(items)
	lq.Wg.Add(work)
	lq.items = append(lq.items, items...)
	lq.m.Unlock()

	for _, item := range items {
		for _, l := range lq.listeners {
			l <- item
		}
	}
}

func (lq *ListQueue[T]) Wait() {
	lq.Wg.Done()
	// Wait for backlogs to be copied to listeners.
	lq.Wg.Wait()
	for _, l := range lq.listeners {
		close(l)
	}
}

// Add takes a key and value, if the key is unique it will add the value to the queue.
func (lq *ListQueue[T]) AddUnique(key string, value T) bool {
	lq.mKeys.RLock()
	_, ok := lq.keys[key]
	lq.mKeys.RUnlock()

	if !ok {
		lq.mKeys.Lock()
		lq.keys[key] = value
		lq.mKeys.Unlock()

		lq.Add(value)
		return true
	} else {
		return false
	}
}
