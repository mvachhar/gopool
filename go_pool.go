package gopool

import "errors"
import "sync/atomic"

type GoPool interface {
	Go(f func())
}

/////////////////////////////////////////////////////
// Dummy GoPool for benchmarking/testing/debugging
type noGoPool struct{}

func (self *noGoPool) Go(f func())      { go f() }
func NewNoPool() (gp GoPool, err error) { gp = new(noGoPool); return }

////////////////////////////////////////////////////
// Real GoPool implementation

type goRtMsg struct {
	work func()
	quit bool
}

func noop() {} //Used to reset goRtMsg
func (self *goRtMsg) reset() {
	self.work = noop
	self.quit = false
}

type goRt struct {
	ch  workChan
	msg goRtMsg //This is buffer space so we don't have to allocate to
	//send work to workers.  This msg cannot be reused
	//until the goRt shows up again on the free worker
	//channel
}

type workChan chan *goRtMsg //Channel type to send workers work
type goRtChan chan *goRt    //Channel type to enqueue/dequeue workers

type GoPoolImpl struct {
	rtCh   goRtChan //Channel with free workers
	reqs   uint64   //Requests for go routines
	misses uint64   //Number of requests that needed new go routine
}

func New(maxSize int) (gp *GoPoolImpl, err error) {
	if maxSize <= 0 {
		err = errors.New("maxSize must be > 0")
		return
	}
	gp = new(GoPoolImpl)
	gp.rtCh = make(goRtChan, maxSize)
	return
}

func (self *GoPoolImpl) Go(f func()) {
	var gr *goRt

	atomic.AddUint64(&self.reqs, 1)
	select {
	case gr = <-self.rtCh:
	default:
		atomic.AddUint64(&self.misses, 1)
		gr = self.createGoRoutine()
	}

	gr.msg.work = f
	gr.msg.quit = false
	gr.ch <- &gr.msg
}

func (self *GoPoolImpl) Requests() uint64 {
	return self.reqs
}

func (self *GoPoolImpl) Misses() uint64 {
	return self.misses
}

func (self *GoPoolImpl) workerMain(gr *goRt) {
	msgCh := gr.ch
	for {
		msg := <-msgCh

		msg.work()
		msg.reset()

		if msg.quit {
			return //Don't repool, we were asked to quit
		}

		select {
		case self.rtCh <- gr: //Return to the pool
		default: //If pool is full, quit go routine
			return
		}
	}
}

func (self *GoPoolImpl) createGoRoutine() (ret *goRt) {
	ret = new(goRt)
	ret.ch = make(workChan)
	ret.msg.reset()
	go self.workerMain(ret)
	return
}
