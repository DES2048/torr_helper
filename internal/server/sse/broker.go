package sse

type Broker[T any] struct {
	stopCh  chan struct{}
	subCh   chan chan T
	unsubCh chan chan T
	msgCh   chan T
}

func NewBroker[T any]() *Broker[T] {
	return &Broker[T]{
		stopCh:  make(chan struct{}),
		subCh:   make(chan chan T, 1),
		unsubCh: make(chan chan T, 1),
		msgCh:   make(chan T, 1),
	}
}

func (b *Broker[T]) Start() {
	subs := make(map[chan T]struct{})

	for {
		select {
		case <-b.stopCh:
			return
		case ch := <-b.subCh:
			subs[ch] = struct{}{}
		case ch := <-b.unsubCh:
			delete(subs, ch)
		case msg := <-b.msgCh:
			for ch := range subs {
				select {
				case ch <- msg:
				default:
				}
			}
		}
	}
}

func (b *Broker[T]) Stop() {
	close(b.stopCh)
}

func (b *Broker[T]) Subscribe(ch chan T) {
	b.subCh <- ch
}

func (b *Broker[T]) Unubscribe(ch chan T) {
	b.unsubCh <- ch
}

func (b *Broker[T]) Pub(msg T) {
	b.msgCh <- msg
}
