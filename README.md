# ListQueue

NewListQueue returns a queue which ensures in order delivery of every past and future event to every subscriber.

	q := lq.NewListQueue[int]()
	q.Add(1, 2)

q.Each() will return a channel which iterates over past and future arguments added with q.Add(...)

	go func() {
		for v := range q.Each() {
	    		// Prints: 	1
	      		//		2
	     		fmt.Println(v)
	        	q.Done()
		}
	
	}()


	q.Add(3, 4)
        q.Close()

	for v := range q.Each() {
		// Prints: 	1
		//		2
		//		3
		//		4
		fmt.Println(v)
		q.Done()
	}



## Package
```
import "github.com/RyanJarv/ListQueue"
```

## CONSTANTS

const ChanSize = 10

## TYPES

### IListQueue[T any]
```
type IListQueue[T any] interface {
    Add(...T)
    Each() chan T
    Close()
    Wait()
}
```

### IWaitGroup interface
```
type IWaitGroup interface {
    Add(i int)
    Done()
    Wait()
}
```

### ListQueue[T any] struct
```
type ListQueue[T any] struct {
// Has unexported fields.
}
```

### func NewListQueue[T any](items ...T) *ListQueue[T]

NewListQueue returns a queue which ensures exactly once, in order delivery of every past and future event to every
subscriber.

```
        q := listQueue.NewListQueue[int]()
        q.Add(1, 2)
        ch := q.Each()
```

q.Each() will return a channel which can be used to iterate over past and future arguments added with Add(...). For
each item returned from this channel. The queue needs to be closed after all items have been added with Add(...) for
loops using q.Each() to exit.

```
           	    go func() {
        	  	      for v := range ch {
        	  	      	    // Prints: 		1
        	  	        	//				2
        	  	        	fmt.Println(v)
        	  	        	q.Done()
        	  	      }
```

Loops created after Close() is called will still work as expected, but new values can't be added with Add(...).

```
        		  q.Add(3, 4)
        		  q.Close()
        		  for v := range q.Each() {
        		  	    // Prints: 		1
        		    	//				2
        		    	//				3
        		    	//				4
        		    	fmt.Println(v)
        		    	q.Done()
        		  }
         	}()
```

### func (s *ListQueue[T]) Add(items ...T)
Add adds all passed arguments to the list and wakes up any paused goroutines to continue reading from the list.

### func (s *ListQueue[T]) Close()

### func (s *ListQueue[T]) Each() chan T

Each returns a channel which will be sent all current and future items passed to Add in the order they were added.
