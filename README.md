# ListQueue

NewListQueue returns a queue which ensures in order delivery of every past and future event to every subscriber.

	q := listQueue.NewListQueue[int]()
	q.Add(1, 2)
	ch := q.Each() // Ensure Each() is called before Wait()

q.Each() will return a channel which iterates over past and future arguments added with q.Add(...). For each item
returned from this channel, q.Done() must be called once the processing of that item is complete.

	go func() {
		for v := range ch {
	    		// Prints: 	1
	      		//		2
	     		fmt.Println(v)
	        	q.Done()
		}
	
		q.Add(3, 4)
	for v := range q.Each() {
		    	// Prints: 	1
	  		//		2
	  		//		3
	  		//		4
	  		fmt.Println(v)
	  		q.Done()
		}
	}()

	// Waits for all events to be delivered to all channels
  	q.Wait()

