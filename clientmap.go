package main

type ClientMap map[*connection]*Client
type ClientIterator <-chan *Client

func (ci ClientIterator) AsSlice() []*Client {
	result := make([]*Client, 0)
	for c := range ci {
		result = append(result, c)
	}
	return result
}

func (cm ClientMap) AllClients() ClientIterator {
	outp := make(chan *Client)
	go func() {
		for _, c := range cm {
			if c != nil {
				outp <- c
			}
		}
		close(outp)
	}()
	return outp
}

func (cm ClientMap) AliveCount() int {
	result := 0
	for _, p := range cm {
		if p != nil && p.kind == Player && p.Alive {
			result++
		}
	}
	return result
}

func (cm ClientMap) PlayerCount() int {
	result := 0
	for _, p := range cm {
		if p != nil && p.kind == Player {
			result++
		}
	}
	return result
}

func (cm ClientMap) Players() ClientIterator {
	outp := make(chan *Client)
	go func() {
		for _, c := range cm {
			if c != nil && c.kind == Player {
				outp <- c
			}
		}
		close(outp)
	}()
	return outp
}

func (cm ClientMap) LivingPlayers() ClientIterator {
	outp := make(chan *Client)
	go func() {
		for _, c := range cm {
			if c != nil && c.kind == Player && c.Alive {
				outp <- c
			}
		}
		close(outp)
	}()
	return outp
}
