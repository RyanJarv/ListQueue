package lq

import (
	"sync"
)

type any interface{}

type IListQueue[T any] interface {
	Add(...T)
}

// NewListQueue returns a queue which ensures in order delivery of every past and future event to every subscriber
//
//    		q := listQueue.NewListQueue[int]()
//    		q.Add(1, 2)
//    		ch := q.Each() // Ensure Each() is called before Wait()
//
// q.Each() will return a channel which iterates over past and future arguments added with q.Add(...). For each itemi
// returned from this channel, q.Done() must be called once the processing of that item is complete.
//
//    	    go func() {
// 	  	    for v := range ch {
// 	  	    	    // Prints: 		1
// 	  	      	//				2
// 	  	      	fmt.Println(v)
// 	  	      	q.Done()
// 	  	    }
//
//   		  q.Add(3, 4)
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
func NewListQueue[T any]() *ListQueue[T] {
	return &ListQueue[T]{
		work:    &sync.WaitGroup{},
		items:   make([]T, 0, 1000),
		c:       sync.NewCond(&sync.Mutex{}),
		notDone: true,
	}
}

type IWaitGroup interface {
	Add(int)
	Done()
	Wait()
}

type ListQueue[T any] struct {
	work        IWaitGroup
	items       []T
	c           *sync.Cond
	notDone     bool
	subscribers int
}

// sendFrom takes a starting index and a channel and sends all items that have been added since the given index
func (s *ListQueue[T]) sendFrom(start int, ch chan T) int {
	var i int
	var v T
	for i, v = range s.items[start:] {
		ch <- v
	}
	return i
}

// next waits for the next batch of items after index i to be placed on the list.
func (s *ListQueue[T]) next(i int) bool {
	s.c.L.Lock()
	s.c.Wait()
	s.c.L.Unlock()
	return s.notDone && len(s.items) != i
}

// Each returns a channel which will be sent all current and future items passed to Add in the order they were added.
func (s *ListQueue[T]) Each() chan T {
	// For each new subscriber all items need to be returned.
	s.subscribers++
	s.work.Add(len(s.items))

	itemCh := make(chan T, 0)
	go func() {
		// Catch this subscriber up to the current tip.
		i := s.sendFrom(0, itemCh)

		// Wait to be notified of more items, then iterate over them.
		for s.next(i) {
			i = s.sendFrom(i, itemCh)
		}
		close(itemCh)
	}()

	return itemCh
}

// Done must be called after processing is finished for each item sent to the channel returned by Each.
func (s *ListQueue[T]) Done() {
	s.work.Done()
}

// Wait must be called after Each() for channels returned by Each() to be closed on completion.
//
// If called before Each() this function will return immediately.
func (s *ListQueue[T]) Wait() {
	s.work.Wait() // Wait for all items to be returned.
	s.notDone = false
	s.c.Broadcast() // Release any goroutines waiting on s.next()
}

// Add adds all passsed arguments to the list and wakes up any paused goroutines to continue reading from the list.
func (s *ListQueue[T]) Add(items ...T) {
	// An item needs to be returned once for each subscriber.
	s.work.Add(len(items) * s.subscribers)
	s.c.L.Lock() // TODO: I don't think we actually need to lock here.
	s.items = append(s.items, items...)
	s.c.Broadcast()
	s.c.L.Unlock()
}
