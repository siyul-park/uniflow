package process

import (
	"context"
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/transaction"
)

// Process is a processing unit that isolates data processing from others.
type Process struct {
	stack        *Stack
	heap         *Heap
	transactions *Transactions
	ctx          context.Context
	done         chan struct{}
	wait         sync.WaitGroup
	mu           sync.Mutex
}

// New creates a new Process.
func New() *Process {
	s := newStack()
	h := newHeap()
	t := newTransactions(s)

	p := &Process{
		stack:        s,
		heap:         h,
		transactions: t,
		done:         make(chan struct{}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.ctx = ctx

	go func() {
		<-p.Done()
		cancel()
	}()

	return p
}

// Stack returns a process's stack.
func (p *Process) Stack() *Stack {
	return p.stack
}

// Heap returns a process's heap.
func (p *Process) Heap() *Heap {
	return p.heap
}

func (p *Process) Transaction(pck *packet.Packet) *transaction.Transaction {
	return p.transactions.Get(pck)
}

func (p *Process) SetTransaction(pck *packet.Packet, tx *transaction.Transaction) {
	p.transactions.Set(pck, tx)
}

// Context returns a process's context.
func (p *Process) Context() context.Context {
	return p.ctx
}

// Done returns a channel that is closed when the process is closed.
func (p *Process) Done() <-chan struct{} {
	return p.done
}

// Ref acquires a lock on the process.
func (p *Process) Ref(count int) {
	p.wait.Add(count)
}

// Close closes the process.
func (p *Process) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.done:
		return
	default:
	}

	p.wait.Wait()
	<-p.stack.Done(nil)

	_ = p.transactions.Commit()

	p.heap.Close()
	p.stack.Close()
	close(p.done)
}
