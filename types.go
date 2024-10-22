package main

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

type Domain struct {
	Name string
	ID   uuid.UUID
}

func (d *Domain) Equal(other *Domain) bool {
	if d.ID == other.ID && d.Name == other.Name {
		return true
	}
	return false
}

type Token struct {
	Ticker   string
	ValueUSD uint64
	Amount   uint64
	Domain   Domain
}

func (t *Token) PartialMatch(other *Token) bool {
	if t.Ticker == other.Ticker && t.ValueUSD == other.ValueUSD {
		return true
	}
	return false
}

func (t *Token) Match(other *Token) bool {
	if t.PartialMatch(other) && t.Domain.Equal(&other.Domain) {
		return true
	}
	return false
}

func (t *Token) Equal(other *Token) bool {
	if t.Match(other) && t.Amount == other.Amount {
		return true
	}
	return false
}

func (t *Token) Validate() error {
	if _, ok := TokenValueMap[t.Ticker]; !ok {
		TokenValueMap[t.Ticker] = t.ValueUSD
		return nil
	}
	if t.ValueUSD != TokenValueMap[t.Ticker] {
		return fmt.Errorf(
			"Token has invalid value got: %d, want %d",
			t.ValueUSD,
			TokenValueMap[t.Ticker],
		)
	}
	return nil
}

type Pool struct {
	TokenA      Token
	TokenB      Token
	VolumeRatio float64
	PriceRatio  float64
	Domain      Domain
}

func (p *Pool) Validate() error {
	if p.TokenA.Match(&p.TokenB) {
		return fmt.Errorf("Invalid Pool Pairing: Tokens are the same type")
	}
	if !p.TokenA.Domain.Equal(&p.Domain) {
		return fmt.Errorf(
			"Invalid Pool: Token (%s) in incorrect domain: got %s, want %s",
			p.TokenA.Ticker,
			p.TokenA.Domain.Name,
			p.Domain.Name,
		)
	}
	if !p.TokenB.Domain.Equal(&p.Domain) {
		return fmt.Errorf(
			"Invalid Pool: Token (%s) in incorrect domain: got %s, want %s",
			p.TokenB.Ticker,
			p.TokenB.Domain.Name,
			p.Domain.Name,
		)
	}
	vr := float64(p.TokenA.Amount) / float64(p.TokenB.Amount)
	if vr != p.VolumeRatio {
		return fmt.Errorf(
			"Invalid Pool: Volume Ratio mismatch: got %f, want %f",
			vr,
			p.VolumeRatio,
		)
	}
	pr := float64(p.TokenA.ValueUSD) / float64(p.TokenB.ValueUSD)
	if pr != p.PriceRatio {
		return fmt.Errorf(
			"Invalid Pool: Price Ratio mismatch: got %f, want %f",
			pr,
			p.PriceRatio,
		)
	}
	return nil
}

type Liquidity struct {
	TokenA Token
	TokenB Token
	Pool   Pool
}

func (l *Liquidity) Validate() error {
	if err := l.Pool.Validate(); err != nil {
		return err
	}
	if (l.TokenA.ValueUSD * l.TokenA.Amount) != (l.TokenB.ValueUSD * l.TokenB.Amount) {
		return fmt.Errorf(
			"Invalid Position: Pair not of equal value: got %d and %d",
			l.TokenA.ValueUSD*l.TokenA.Amount,
			l.TokenB.ValueUSD*l.TokenB.Amount,
		)
	}
	if !l.TokenA.Match(&l.Pool.TokenA) {
		return fmt.Errorf(
			"Invalid Position: Token A not the same as Pool's Token A: got %s want %s",
			l.TokenA.Ticker,
			l.Pool.TokenA.Ticker,
		)
	}
	if l.TokenA.Amount > l.Pool.TokenA.Amount {
		return fmt.Errorf(
			"Invalid Position: Token (%s) has greater volume than the pool's capacity: got %d, pool has %d",
			l.TokenA.Ticker,
			l.TokenA.Amount,
			l.Pool.TokenA.Amount,
		)
	}
	if l.TokenA.ValueUSD > l.Pool.TokenA.ValueUSD {
		return fmt.Errorf(
			"Invalid Position: Token (%s) has greater value than the pool: got %d, pool has %d",
			l.TokenA.Ticker,
			l.TokenA.ValueUSD,
			l.Pool.TokenA.ValueUSD,
		)
	}
	if !l.TokenB.Match(&l.Pool.TokenB) {
		return fmt.Errorf(
			"Invalid Position: Token B not the same as Pool's Token B: got %s want %s",
			l.TokenB.Ticker,
			l.Pool.TokenB.Ticker,
		)
	}
	if l.TokenB.Amount > l.Pool.TokenB.Amount {
		return fmt.Errorf(
			"Invalid Position: Token (%s) has greater volume than the pool's capacity: got %d, pool has %d",
			l.TokenB.Ticker,
			l.TokenB.Amount,
			l.Pool.TokenB.Amount,
		)
	}
	if l.TokenB.ValueUSD > l.Pool.TokenB.ValueUSD {
		return fmt.Errorf(
			"Invalid Position: Token (%s) has greater value than the pool: got %d, pool has %d",
			l.TokenB.Ticker,
			l.TokenB.ValueUSD,
			l.Pool.TokenB.ValueUSD,
		)
	}
	return nil
}

type Solver struct {
	Name      string
	ID        uuid.UUID
	Positions []Liquidity
	OrderChan <-chan *Order
	CloseFunc func()
}

type User struct {
	Name   string
	ID     uuid.UUID
	Tokens []Token
	Orders []*Order
}

type Bid struct {
	OrderID      uuid.UUID
	SolverID     uuid.UUID
	UserID       uuid.UUID
	OrderChannel chan *Bid
	TargetToken  Token
	AmountUsed   uint64
	Path         []Liquidity
	Accepted     bool
}

type Order struct {
	// OrderID ia the UUID representing this specific order.
	OrderID uuid.UUID
	// UserID is the UUID of the user posting the Order.
	UserID uuid.UUID
	// SolverID is an ID of the Solver used to fulfil the Order.
	SolverID uuid.UUID
	// BidChannel is a channel used by the solvers to send their bids for the
	// order if they can fill it, the best bid(s) will be used to fill the order.
	BidChannel chan *Bid
	//  OriginToken is the Token being offered in the swap for the TargetToken
	OriginToken Token
	// TargetToken is the desired token in the order swapped with the OriginToken
	TargetToken Token
	// EntryTime is the time at which the order was posted, when the order was
	// cr eated by the user.
	EntryTime time.Time
	// ExitTime is the time at which the order was exited (cancelled/fulfilled)
	// this timestamp doesn't guarentee the order was filled, just that it is closed.
	ExitTime *time.Time
	// TimeoutPeriod is the duration from the EntryTime that the order is open for.
	// Once the timeout period has passed, the order is cancelled and no funds are
	// exchanged. This puts a limit on the order and can be configured on creation.
	TimeoutPeriod time.Duration
	// BidPath is the path the filled order takes to be filled via the solver's
	// different liquidity positions in order to settle with the full TargetToken
	// amount on the correct domain. This is only set once the order has been
	// filled by an accepted bid.
	BidPath []Liquidity
	// Filled is a boolean value showing whether the order has been fully solved.
	Filled bool
}

func (o *Order) Validate() error {
	if o.ExitTime != nil {
		if !o.EntryTime.Before(*o.ExitTime) {
			return fmt.Errorf(
				"Invalid Order: Order Exit Time is before order Entry Time: entry %v, exit %v",
				o.EntryTime,
				*o.ExitTime,
			)
		}
		if o.EntryTime.Add(o.TimeoutPeriod).Before(*o.ExitTime) {
			return fmt.Errorf("Invalid Order: Exit Time is after Timeout Period has elapsed")
		}
	}
	if o.Filled {
		if len(o.BidPath) < 1 {
			return fmt.Errorf("Invalid Order: Filled but no solved path found")
		}
		if !o.BidPath[len(o.BidPath)-1].TokenB.Match(&o.TargetToken) {
			return fmt.Errorf(
				"Invalid Order: Final token from solved path mismatch: got %v, want %v",
				o.BidPath[len(o.BidPath)-1].TokenB,
				o.TargetToken,
			)
		}
		if o.SolverID.String() == "00000000-0000-0000-0000-000000000000" {
			return fmt.Errorf("Invalid Order: Filled order has no matching Solver ID")
		}
	}
	if !o.Filled &&
		(len(o.BidPath) > 0 || o.SolverID.String() != "00000000-0000-0000-0000-000000000000") {
		return fmt.Errorf(
			"Invalid Order: Filled fields present but not market as filled",
		)
	}
	return nil
}

func (u *User) CreateOrder(origin, target Token, timeout time.Duration) (*Order, error) {
	if err := uuid.Validate(u.ID.String()); err != nil {
		return nil, err
	}
	order := &Order{
		OrderID:     uuid.New(),
		UserID:      u.ID,
		BidChannel:  make(chan *Bid),
		OriginToken: origin,
		TargetToken: target,
	}
	if timeout > 0 {
		order.TimeoutPeriod = timeout
	} else {
		// No timeout / 290 years
		order.TimeoutPeriod = time.Duration(math.MaxInt64)
	}
	order.EntryTime = time.Now()
	if err := order.Validate(); err != nil {
		return nil, err
	}
	return order, nil
}
