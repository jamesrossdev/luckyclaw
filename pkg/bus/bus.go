package bus

import (
	"context"
	"fmt"
	"sync"
)

type MessageBus struct {
	inbound     chan InboundMessage
	outbound    chan OutboundMessage
	handlers    map[string]MessageHandler
	outHandlers map[string]OutboundHandler
	closed      bool
	mu          sync.RWMutex
}

func NewMessageBus() *MessageBus {
	return &MessageBus{
		inbound:     make(chan InboundMessage, 100),
		outbound:    make(chan OutboundMessage, 100),
		handlers:    make(map[string]MessageHandler),
		outHandlers: make(map[string]OutboundHandler),
	}
}

func (mb *MessageBus) PublishInbound(msg InboundMessage) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	if mb.closed {
		return
	}
	mb.inbound <- msg
}

func (mb *MessageBus) ConsumeInbound(ctx context.Context) (InboundMessage, bool) {
	select {
	case msg := <-mb.inbound:
		return msg, true
	case <-ctx.Done():
		return InboundMessage{}, false
	}
}

func (mb *MessageBus) PublishOutbound(msg OutboundMessage) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	if mb.closed {
		return
	}
	mb.outbound <- msg
}

// SendDirect sends a message synchronously through the registered outbound handler,
// returning any error (e.g. Discord 403). Use this when the caller needs to
// know whether the send succeeded (e.g. the message tool).
func (mb *MessageBus) SendDirect(ctx context.Context, msg OutboundMessage) error {
	mb.mu.RLock()
	handler, ok := mb.outHandlers[msg.Channel]
	mb.mu.RUnlock()
	if !ok {
		return fmt.Errorf("no outbound handler for channel: %s", msg.Channel)
	}
	return handler(ctx, msg)
}

// RegisterOutboundHandler registers a synchronous send handler for a channel.
// Used by the channel manager so the message tool can do synchronous sends.
func (mb *MessageBus) RegisterOutboundHandler(channel string, handler OutboundHandler) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.outHandlers[channel] = handler
}

func (mb *MessageBus) SubscribeOutbound(ctx context.Context) (OutboundMessage, bool) {
	select {
	case msg := <-mb.outbound:
		return msg, true
	case <-ctx.Done():
		return OutboundMessage{}, false
	}
}

func (mb *MessageBus) RegisterHandler(channel string, handler MessageHandler) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.handlers[channel] = handler
}

func (mb *MessageBus) GetHandler(channel string) (MessageHandler, bool) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	handler, ok := mb.handlers[channel]
	return handler, ok
}

func (mb *MessageBus) Close() {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	if mb.closed {
		return
	}
	mb.closed = true
	close(mb.inbound)
	close(mb.outbound)
}
