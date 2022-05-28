package backend

type Broker struct {
	stopCh    chan bool
	publishCh chan Pixel
	subCh     chan chan Pixel
	unsubCh   chan chan Pixel
}

func NewBroker() *Broker {
	return &Broker{
		stopCh:    make(chan bool),
		publishCh: make(chan Pixel, 100),
		subCh:     make(chan chan Pixel, 10),
		unsubCh:   make(chan chan Pixel, 10),
	}
}

func (b *Broker) Start() {
	subs := map[chan Pixel]bool{}
	for {
		select {
		case <-b.stopCh:
			return
		case msgCh := <-b.subCh:
			subs[msgCh] = true
		case msgCh := <-b.unsubCh:
			delete(subs, msgCh)
		case msg := <-b.publishCh:
			for msgCh := range subs {
				// msgCh is buffered, use non-blocking send to protect the broker:
				select {
				case msgCh <- msg:
				default:
				}
			}
		}
	}
}

func (b *Broker) Stop() {
	close(b.stopCh)
}

func (b *Broker) Subscribe() chan Pixel {
	msgCh := make(chan Pixel, 5)
	b.subCh <- msgCh
	return msgCh
}

func (b *Broker) Unsubscribe(msgCh chan Pixel) {
	b.unsubCh <- msgCh
}

func (b *Broker) Publish(msg Pixel) {
	b.publishCh <- msg
}
