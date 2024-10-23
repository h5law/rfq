# rfq

A simple \[**R**\]equest \[**F**\]or \[**Q**\]uote cross-domain, auction-styled,
order-bid pairing system for Market Makers / Solvers to respond to User's posted
orders with a bid specifying how through their positions in liquidity pools an
order to swap token `A` for token `B` on domains `X` and `Y` respectively can be
fulfilled in the shortest number of steps (swaps in a pool).

This is simply a PoC for such a system and makes many assumptions, and should not
be used outside of testing or simple modelling. For a more detailed insight into
this system look into [doc.md](./doc.md) which specifies much more about the
design, architecture and overall function of the system as well as its pros and cons.

## Build and Run

In order to build and run the simulation simply run the following:

```console
$ go build
$ ./rfq
User: bob - Invalid Order: User does not have or have enough of token: TOK3-dom1
User: alice - Invalid Order: User does not have or have enough of token: TOK3-dom1
alice Posting Order: TOK1-dom1 -> TOK3-dom1 [5503c1e9-193a-48c9-acc0-e6c845de680b]...
Order 5503c1e9-193a-48c9-acc0-e6c845de680b: Filled by Solver 44ad58ee-6d6b-4f81-9707-254a790ee74e
	Path: [TOK1-dom1 -> TOK3-dom1] ✅
alice Posting Order: TOK1-dom1 -> TOK2-dom1 [bf3c092a-c5c4-433c-a600-fe61a1938469]...
Order bf3c092a-c5c4-433c-a600-fe61a1938469: Filled by Solver 44ad58ee-6d6b-4f81-9707-254a790ee74e
	Path: [TOK1-dom1 -> TOK2-dom1] ✅
bob Posting Order: TOK1-dom0 -> TOK1-dom1 [f1ef17fc-75a6-4551-8609-e84d41f44182]...
Order f1ef17fc-75a6-4551-8609-e84d41f44182: Filled by Solver f2e9df4d-d29c-4766-b26c-5c404ce829ad
	Path: [TOK1-dom0 -> TOK2-dom0] -> [TOK2-dom1 -> TOK1-dom1] ✅
bob Posting Order: TOK1-dom1 -> TOK2-dom0 [284b963a-38ae-42de-8a84-52d46cd0bf2e]...
Order 284b963a-38ae-42de-8a84-52d46cd0bf2e: Filled by Solver f2e9df4d-d29c-4766-b26c-5c404ce829ad
	Path: [TOK1-dom0 -> TOK2-dom0] ✅
bob Posting Order: TOK2-dom0 -> TOK3-dom0 [61d02ab1-2319-4e50-88c0-46698fd6344a]...
Order 61d02ab1-2319-4e50-88c0-46698fd6344a: not filled ❌
bob Posting Order: TOK2-dom0 -> TOK3-dom1 [36028cb1-035d-481b-8386-6847eddda963]...
Order 36028cb1-035d-481b-8386-6847eddda963: not filled ❌
```

In order to test new paths and orders simply edit the [./config.yaml](./config.yaml)
file and add new orders, users, and solvers to the environment.
