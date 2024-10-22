package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	yaml "gopkg.in/yaml.v2"
)

var (
	DomainMap     map[string]uuid.UUID
	TokenValueMap map[string]uint64
	OrderList     []*Order
)

func init() {
	DomainMap = make(map[string]uuid.UUID)
	TokenValueMap = make(map[string]uint64)
	OrderList = make([]*Order, 0, 16)
}

type DomainConfig struct {
	Name string `yaml:"name"`
}

type TokenConfig struct {
	Ticker   string       `yaml:"ticker"`
	ValueUSD uint64       `yaml:"usd"`
	Amount   uint64       `yaml:"amount"`
	Domain   DomainConfig `yaml:"domain"`
}

type UserConfig struct {
	Name   string        `yaml:"name"`
	Tokens []TokenConfig `yaml:"tokens"`
}

type PoolConfig struct {
	Pair   [2]TokenConfig `yaml:"pair"`
	Domain DomainConfig   `yaml:"domain"`
}

type PositionConfig struct {
	Pair [2]TokenConfig `yaml:"pair"`
	Pool PoolConfig     `yaml:"pool"`
}

type SolverConfig struct {
	Name      string           `yaml:"name"`
	Positions []PositionConfig `yaml:"positions"`
}

type OrderConfig struct {
	User        string      `yaml:"user"`
	OriginToken TokenConfig `yaml:"origin"`
	TargetToken TokenConfig `yaml:"target"`
	Timeout     int64       `yaml:"timeout"`
}

type AgentsConfigs struct {
	Users   []UserConfig   `yaml:"users"`
	Solvers []SolverConfig `yaml:"solvers"`
	Orders  []OrderConfig  `yaml:"orders"`
}

func ReadConfigFile(path string, conf *AgentsConfigs) error {
	file, err := os.OpenFile(path, os.O_RDONLY, 0o644)
	if err != nil {
		return err
	}
	configBz, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(configBz, conf); err != nil {
		return err
	}
	return nil
}

func convertTokens(tokensConf []TokenConfig) []Token {
	tokens := make([]Token, len(tokensConf))
	for j, t := range tokensConf {
		if _, ok := DomainMap[t.Domain.Name]; !ok {
			DomainMap[t.Domain.Name] = uuid.New()
		}
		tokens[j] = Token{
			Ticker:   t.Ticker,
			ValueUSD: t.ValueUSD,
			Amount:   t.Amount,
			Domain: Domain{
				Name: t.Domain.Name,
				ID:   DomainMap[t.Domain.Name],
			},
		}
		if err := tokens[j].Validate(); err != nil {
			panic(err)
		}
	}
	return tokens
}

func (a *AgentsConfigs) ConvertUsers() map[uuid.UUID]*User {
	if len(a.Users) == 0 {
		return nil
	}
	users := make(map[uuid.UUID]*User, len(a.Users))
	for _, u := range a.Users {
		tokens := convertTokens(u.Tokens)
		id := uuid.New()
		users[id] = &User{
			Name:   u.Name,
			ID:     id,
			Tokens: tokens,
		}
	}
	return users
}

func (a *AgentsConfigs) ConvertSolvers() map[uuid.UUID]*Solver {
	if len(a.Solvers) == 0 {
		return nil
	}
	solvers := make(map[uuid.UUID]*Solver, len(a.Solvers))
	for _, s := range a.Solvers {
		id := uuid.New()
		solver := &Solver{
			Name:      s.Name,
			ID:        id,
			Positions: []Liquidity{},
		}
		if len(s.Positions) == 0 {
			solvers[id] = solver
			continue
		}
		solverPositions := make([]Liquidity, len(s.Positions))
		for j, p := range s.Positions {
			if len(p.Pair) != 2 {
				panic(fmt.Sprintf("Invalid Liquidity Pairing: got %d tokens want 2", len(p.Pair)))
			}
			if len(p.Pool.Pair) != 2 {
				panic(
					fmt.Sprintf(
						"Invalid Liquidity Pairing: got %d tokens want 2",
						len(p.Pool.Pair),
					),
				)
			}
			tokensAB := convertTokens(p.Pair[:])
			solverPositions[j].TokenA = tokensAB[0]
			solverPositions[j].TokenB = tokensAB[1]
			tokensAB = convertTokens(p.Pool.Pair[:])
			solverPositions[j].Pool = Pool{
				TokenA:      tokensAB[0],
				TokenB:      tokensAB[1],
				PriceRatio:  float64(tokensAB[0].ValueUSD) / float64(tokensAB[1].ValueUSD),
				VolumeRatio: float64(tokensAB[0].Amount) / float64(tokensAB[1].Amount),
				Domain: Domain{
					Name: p.Pool.Domain.Name,
					ID:   DomainMap[p.Pool.Domain.Name],
				},
			}
			if err := solverPositions[j].Pool.Validate(); err != nil {
				panic(err)
			}
		}
		solver.Positions = solverPositions
		solvers[id] = solver
	}
	return solvers
}

func (ac *AgentsConfigs) ConvertOrders(users map[uuid.UUID]*User) []*Order {
	for _, order := range ac.Orders {
		origin := convertTokens([]TokenConfig{order.OriginToken})
		target := convertTokens([]TokenConfig{order.TargetToken})
		timeout := order.Timeout
		for _, user := range users {
			if user.Name != order.User {
				continue
			}
			var ord *Order
			var err error
			if timeout <= 0 {
				ord, err = user.CreateOrder(origin[0], target[0], 0)
			} else {
				ord, err = user.CreateOrder(origin[0], target[0], time.Duration(timeout))
			}
			if err != nil {
				panic(err)
			}
			OrderList = append(OrderList, ord)
			break
		}
	}
	return OrderList
}
