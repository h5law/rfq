package main

import (
	"fmt"
	"sync"
	"time"
)

type Auction struct {
	orders []chan *Order
	mu     *sync.RWMutex
}

func NewAuction() *Auction {
	return &Auction{
		orders: make([]chan *Order, 0, 128),
		mu:     new(sync.RWMutex),
	}
}

func (p *Auction) Orders() (<-chan *Order, func()) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan *Order)
	p.orders = append(p.orders, ch)

	return ch, func() {
		p.mu.Lock()
		defer p.mu.Unlock()

		for i, c := range p.orders {
			if c == ch {
				p.orders = append(p.orders[:i], p.orders[i+1:]...)
				close(ch)
				return
			}
		}
	}
}

func (p *Auction) Post(order *Order) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, ch := range p.orders {
		ch <- order
	}
}

func (s *Solver) Start(auction *Auction) {
	orderChan, closerFn := auction.Orders()
	s.OrderChan = orderChan
	s.CloseFunc = closerFn

	go func() {
		for {
			order, ok := <-s.OrderChan
			if !ok {
				return
			}
			s.CheckBid(order)
		}
	}()
}

func (s *Solver) Stop() {
	s.CloseFunc()
}

func (u *User) Start(auction *Auction) {
	for _, order := range u.Orders {
		if order == nil {
			continue
		}
		bids := make([]*Bid, 0, len(solvers))
		fmt.Printf("%s Posting Order: %s-%s -> %s-%s [%s]...\n",
			UserIDMap[order.UserID].Name,
			order.OriginToken.Ticker, order.OriginToken.Domain.Name,
			order.TargetToken.Ticker, order.TargetToken.Domain.Name,
			order.OrderID.String())
		auction.Post(order)
	inner:
		for {
			select {
			case bid, ok := <-order.BidChannel:
				if !ok {
					break inner
				}
				if bid == nil {
					bids = append(bids, &Bid{})
				} else {
					bids = append(bids, bid)
				}
			default:
				if time.Now().After(order.EntryTime.Add(order.TimeoutPeriod)) {
					break inner
				}
				if len(bids) == len(solvers) {
					break inner
				}
			}
		}
		close(order.BidChannel) // all bids in, no more needed
		u.AcceptBid(order, bids)
		if err := order.Validate(); err != nil {
			panic(err)
		}
		if !order.Filled {
			fmt.Printf("Order %s: not filled ❌\n", order.OrderID.String())
			continue
		}
		fmt.Printf(
			"Order %s: Filled by Solver %s\n\tPath: ",
			order.OrderID.String(),
			order.SolverID.String(),
		)
		for i, lp := range order.BidPath {
			if i < len(order.BidPath)-1 {
				fmt.Printf("[%s-%s -> %s-%s] -> ",
					lp.TokenA.Ticker, lp.TokenA.Domain.Name,
					lp.TokenB.Ticker, lp.TokenB.Domain.Name,
				)
				continue
			}
			fmt.Printf("[%s-%s -> %s-%s] ✅\n",
				lp.TokenA.Ticker, lp.TokenA.Domain.Name,
				lp.TokenB.Ticker, lp.TokenB.Domain.Name,
			)
		}
	}
}
