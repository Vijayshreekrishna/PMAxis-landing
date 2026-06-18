package ws

import (
	"context"
	"strings"

	"github.com/pmaxis/pmaxis/libs/logger"
)

// broadcastMsg carries a message and an optional market scope.
// marketID == "" means broadcast to every connected client.
type broadcastMsg struct {
	data     []byte
	marketID string
}

// subscribeCmd updates a client's market subscription set.
// Empty markets slice means subscribe to everything.
type subscribeCmd struct {
	client  *Client
	markets []string
}

// walletBroadcastMsg carries a wallet activity message scoped to a single address.
type walletBroadcastMsg struct {
	data   []byte
	wallet string // lowercase address
}

// walletSubCmd updates a client's wallet subscription set.
type walletSubCmd struct {
	client  *Client
	wallets []string
}

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	clients         map[*Client]bool
	broadcast       chan broadcastMsg
	register        chan *Client
	unregister      chan *Client
	subscribe       chan subscribeCmd
	walletBroadcast chan walletBroadcastMsg
	subscribeWallet chan walletSubCmd
	log             logger.Interface
}

func NewHub(log logger.Interface) *Hub {
	return &Hub{
		broadcast:       make(chan broadcastMsg, 1024),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		subscribe:       make(chan subscribeCmd, 64),
		walletBroadcast: make(chan walletBroadcastMsg, 256),
		subscribeWallet: make(chan walletSubCmd, 64),
		clients:         make(map[*Client]bool),
		log:             log.With("component", "ws_hub"),
	}
}

func (h *Hub) Run(ctx context.Context) {
	h.log.Info("ws hub starting")
	for {
		select {
		case <-ctx.Done():
			h.log.Info("ws hub shutting down")
			return
		case client := <-h.register:
			h.clients[client] = true
			h.log.Info("client connected", "active_clients", len(h.clients))
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.log.Info("client disconnected", "active_clients", len(h.clients))
			}
		case sub := <-h.subscribe:
			if _, ok := h.clients[sub.client]; ok {
				sub.client.subscriptions = make(map[string]bool, len(sub.markets))
				for _, m := range sub.markets {
					if m != "" {
						sub.client.subscriptions[m] = true
					}
				}
				h.log.Debug("client subscribed", "markets", sub.markets)
			}
		case sub := <-h.subscribeWallet:
			if _, ok := h.clients[sub.client]; ok {
				sub.client.walletSubs = make(map[string]bool, len(sub.wallets))
				for _, w := range sub.wallets {
					if w != "" {
						sub.client.walletSubs[strings.ToLower(w)] = true
					}
				}
				h.log.Debug("client subscribed to wallets", "wallets", sub.wallets)
			}
		case msg := <-h.walletBroadcast:
			for client := range h.clients {
				if !client.walletSubs[msg.wallet] {
					continue
				}
				select {
				case client.send <- msg.data:
				default:
					close(client.send)
					delete(h.clients, client)
					h.log.Warn("dropped client due to slow buffer")
				}
			}
		case msg := <-h.broadcast:
			for client := range h.clients {
				// If the client has subscriptions and this message is scoped to a
				// specific market, skip clients not subscribed to that market.
				if len(client.subscriptions) > 0 && msg.marketID != "" && !client.subscriptions[msg.marketID] {
					continue
				}
				select {
				case client.send <- msg.data:
				default:
					close(client.send)
					delete(h.clients, client)
					h.log.Warn("dropped client due to slow buffer")
				}
			}
		}
	}
}

// Broadcast sends data to all connected clients (no market filter).
func (h *Hub) Broadcast(msg []byte) {
	h.broadcast <- broadcastMsg{data: msg}
}

// BroadcastToMarket sends data only to clients subscribed to marketID,
// plus clients with no active subscription (receive everything).
func (h *Hub) BroadcastToMarket(data []byte, marketID string) {
	h.broadcast <- broadcastMsg{data: data, marketID: marketID}
}

// Subscribe updates a client's market filter set.
func (h *Hub) Subscribe(c *Client, markets []string) {
	h.subscribe <- subscribeCmd{client: c, markets: markets}
}

// SubscribeWallet updates a client's wallet filter set.
func (h *Hub) SubscribeWallet(c *Client, wallets []string) {
	h.subscribeWallet <- walletSubCmd{client: c, wallets: wallets}
}

// BroadcastToWallet sends data only to clients subscribed to the given wallet address.
func (h *Hub) BroadcastToWallet(data []byte, wallet string) {
	h.walletBroadcast <- walletBroadcastMsg{data: data, wallet: strings.ToLower(wallet)}
}
