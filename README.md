# rfq

A simple **\[R\]**equest **\[F\]**or **\[Q\]**uote cross-domain, auction-styled,
order-bid pairing system for Market Makers / Solvers to respond to User's posted
orders with a bid specifying how through their positions in liquidity pools an
order to swap token `A` for token `B` on domains `X` and `Y` respectively can be
fulfilled in the shortest number of steps (swaps in a pool).

This is simply a PoC for such a system and makes many assumptions, and should not
be used outside of testing or simple modelling. For a more detailed insight into
this system look into [doc.md](./doc.md) which specifies much more about the
design, architecture and overall function of the system as well as its pros and cons.
