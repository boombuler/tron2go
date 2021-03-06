package main

func BufferChannel(in <-chan []byte, next chan<- []byte) {
	defer close(next)

	// pending events (this is the "infinite" part)
	pending := [][]byte{}

recv:
	for {
		// Ensure that pending always has values so the select can
		// multiplex between the receiver and sender properly
		if len(pending) == 0 {
			v, ok := <-in
			if !ok {
				// in is closed, flush values
				break
			}
			// We now have something to send
			pending = append(pending, v)
		}

		select {
		// Queue incoming values
		case v, ok := <-in:
			if !ok {
				// in is closed, flush values
				break recv
			}
			pending = append(pending, v)

		// Send queued values
		case next <- pending[0]:
			pending = pending[1:]
		}
	}

	// After in is closed, we may still have events to send
	for _, v := range pending {
		next <- v
	}
}
