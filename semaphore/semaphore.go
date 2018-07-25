package semaphore

import "golang.org/x/sync/errgroup"

type Semaphore struct {
	eg *errgroup.Group
	ch chan bool
}

func NewSemaphore(n int) *Semaphore {
	return &Semaphore{
		eg: &errgroup.Group{},
		ch: make(chan bool, n),
	}
}

func (p *Semaphore) Go(f func() error) {
	p.eg.Go(func() error {
		defer func() {
			<-p.ch
		}()
		p.ch <- true
		return f()
	})
}

func (p *Semaphore) Wait() error {
	return p.eg.Wait()
}
