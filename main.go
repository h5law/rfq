package main

import (
	"time"

	"github.com/google/uuid"
)

var (
	users   map[uuid.UUID]*User
	solvers map[uuid.UUID]*Solver
	orders  []*Order
)

func init() {
	users = make(map[uuid.UUID]*User)
	solvers = make(map[uuid.UUID]*Solver)
	orders = make([]*Order, 0, 16)
}

func main() {
	var ac AgentsConfigs
	if err := ReadConfigFile("./config.yaml", &ac); err != nil {
		panic(err)
	}
	auction := NewAuction()
	users = ac.ConvertUsers()
	solvers = ac.ConvertSolvers()
	orders = ac.ConvertOrders(users)
	for _, s := range solvers {
		s.Start(auction)
		defer s.Stop()
	}
	time.Sleep(10 * time.Millisecond)
	for _, ord := range orders {
		user, ok := users[ord.UserID]
		if !ok {
			panic("User not found for order")
		}
		user.Orders = append(user.Orders, ord)
	}
	for _, u := range users {
		u.Start(auction)
	}
}
