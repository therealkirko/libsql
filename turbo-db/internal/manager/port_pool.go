package manager

import (
	"fmt"
	"sync"
)

// PortPool manages a pool of available ports
type PortPool struct {
	start     int
	end       int
	allocated map[int]bool
	mu        sync.Mutex
}

// NewPortPool creates a new port pool
func NewPortPool(start, end int) *PortPool {
	return &PortPool{
		start:     start,
		end:       end,
		allocated: make(map[int]bool),
	}
}

// Allocate allocates an available port
func (p *PortPool) Allocate() (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for port := p.start; port < p.end; port++ {
		if !p.allocated[port] {
			p.allocated[port] = true
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports in range %d-%d", p.start, p.end)
}

// Release releases a port back to the pool
func (p *PortPool) Release(port int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.allocated, port)
}

// MarkAllocated marks a port as allocated (used during restoration)
func (p *PortPool) MarkAllocated(port int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.allocated[port] = true
}

// IsAllocated checks if a port is allocated
func (p *PortPool) IsAllocated(port int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.allocated[port]
}
