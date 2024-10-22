package main

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"time"
)

type Trade struct {
	TokenA Token
	TokenB Token
}

type QueueElement struct {
	Token Token
	Path  []Trade
}

func (s *Solver) findOrderPath(order *Order) []Liquidity {
	graph := make(map[Token][]Token)
	for _, p := range s.Positions {
		graph[p.TokenA] = append(graph[p.TokenA], p.TokenB)
	}

	ov := order.OriginToken.Amount * order.OriginToken.ValueUSD
	tv := order.TargetToken.Amount * order.TargetToken.ValueUSD
	queue := []QueueElement{{
		Token: order.OriginToken,
		Path:  []Trade{},
	}}
	visited := make(map[string]bool)
	reducedGraph := make(map[string][]Token, len(graph))
	for neighbour, neighbours := range graph {
		reducedGraph[neighbour.Ticker+"-"+neighbour.Domain.Name] = append(
			reducedGraph[neighbour.Ticker+"-"+neighbour.Domain.Name],
			neighbours...)
	}

	path := []Trade{}
	for len(queue) > 0 {
		elem := queue[0]
		queue = queue[1:]

		if elem.Token.Match(&order.TargetToken) {
			path = elem.Path
			break
		}

		if !visited[elem.Token.Ticker+"-"+elem.Token.Domain.Name] {
			visited[elem.Token.Ticker+"-"+elem.Token.Domain.Name] = true
			for _, neighbour := range reducedGraph[elem.Token.Ticker+"-"+elem.Token.Domain.Name] {
				if !visited[neighbour.Ticker+"-"+neighbour.Domain.Name] {
					if ov > elem.Token.Amount*elem.Token.ValueUSD ||
						tv > elem.Token.Amount*elem.Token.ValueUSD {
						continue
					}
					newPath := append([]Trade{}, elem.Path...)
					newPath = append(newPath, Trade{elem.Token, neighbour})
					queue = append(queue, QueueElement{Token: neighbour, Path: newPath})
				}
			}

			for a, bs := range graph {
				if elem.Token.PartialMatch(&a) && !elem.Token.Domain.Equal(&a.Domain) {
					newElem := a.Ticker + "-" + a.Domain.Name
					for _, b := range bs {
						if !visited[newElem] {
							if ov >= a.Amount*a.ValueUSD || tv >= a.Amount*a.ValueUSD {
								continue
							}
							newPath := append([]Trade{}, elem.Path...)
							newPath = append(newPath, Trade{a, b})
							queue = append(queue, QueueElement{Token: b, Path: newPath})
						}
					}
				}
			}
		}
	}

	if len(path) == 0 {
		return nil
	}

	trades := make([]Liquidity, len(path))
	for i, p := range path {
		tokenA := p.TokenA
		tokenB := p.TokenB
		for _, pos := range s.Positions {
			if pos.TokenA.Match(&tokenA) && pos.TokenB.Match(&tokenB) {
				trades[i] = pos
				break
			}
		}
	}

	return trades
}

func (s *Solver) CheckBid(order *Order) {
	path := s.findOrderPath(order)
	if path == nil {
		order.BidChannel <- nil
		return
	}
	order.BidChannel <- &Bid{
		OrderID:      order.OrderID,
		UserID:       order.UserID,
		SolverID:     s.ID,
		TargetToken:  order.TargetToken,
		AmountUsed:   order.TargetToken.Amount,
		OrderChannel: order.BidChannel,
		Path:         path,
	}
}

func (u *User) AcceptBid(order *Order, bids []*Bid) {
	validBids := make([]*Bid, 0, len(bids))
Outer:
	for _, b := range bids {
		if b == nil {
			continue
		}
		if len(b.Path) == 0 {
			continue
		}
		for _, p := range b.Path {
			if err := p.Validate(); err != nil {
				fmt.Fprintf(os.Stderr, "Invalid pool in bid path: %v", err)
				continue Outer
			}
		}
		validBids = append(validBids, b)
	}
	if len(validBids) == 0 {
		return
	}
	slices.SortFunc(validBids, func(a, b *Bid) int {
		return cmp.Compare(len(a.Path), len(b.Path))
	})
	chosen := validBids[0]
	now := time.Now()
	order.Filled = true
	chosen.Accepted = true
	order.ExitTime = &now
	order.SolverID = chosen.SolverID
	order.BidPath = chosen.Path
}
