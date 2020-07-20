package conn

import (
	"net"
	"sync"
)

type Spool struct {
	mu   sync.Mutex
	Pool map[string]net.Conn // {client addr: *net.conn}
}

func NewConnPool() *Spool {
	return &Spool{Pool: map[string]net.Conn{}}
}

func (p *Spool) Put(c net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.Pool[c.RemoteAddr().String()]; ok {
		p.Remove(c.RemoteAddr().String())
	}
	p.Pool[c.LocalAddr().String()] = c
}

func (p *Spool) Load(addr string) net.Conn {
	p.mu.Lock()
	defer p.mu.Unlock()
	if v, ok := p.Pool[addr]; ok {
		return v
	}
	return nil
}

func (p *Spool) Remove(addr string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.Pool[addr]; ok {
		delete(p.Pool, addr)
	}
}

func (p *Spool) Len() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.Pool)
}

func (p *Spool) WipeOut(c net.Conn) {
	p.Remove(c.RemoteAddr().String())
	c.Close()
}
