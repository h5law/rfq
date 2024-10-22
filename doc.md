# **R**equest **F**or **Q**uote

This application is a basic PoC of a cross-domain RFQ auction styled order-bid
pairing system for users and market-makers / solvers to interact with each other
efficiently in an automated manner.

The auction styled order-bid pairing system is able to find, for a specific
order, the MM / Solver and the specific path through the MM's positions in
liquidity pools across multiple domains such that the order is filled in the
shortest number of steps as is possible. However, there are many assumptions and
limitations in this implementation - which is why it is only a PoC and not a
production ready (or usable in any manner besides testing) implementation.

** This is only a PoC and is not intended to be used outside of testing at all.**

## Features, Components, Assumptions and Details

- Agents
  - Agents are defined as typed units in the codebase such that they interact
    with the environment and are used to fulfil the desired use-case of the
    system when combined. They are defined as follows:
    - Tokens and Domains
      - Tokens and Domains are abstractions to represent coins and chains in the
        simplest sense. Tokens exist on any given Domain and may be inside of a 
        liquidity pool in the case of MMs / Solvers. They have a specific value
        in USD as well as amount held (or deposited in the case of a pool) as
        well as ticker defined in addition to the Domain. Domains are simply a 
        pairing of a domain name (e.g. "dom0") and a generated UUID. Their names
        are unique in this implementation and for each unique name their is a 
        corresponding unique UUID per domain.
        - Token's USD value is represented of a single token not the entire of
          the amount held.
        - Token's Ticker values are assumed to be unique.
        - Token's are assumed to be Partially Equal; a Match; or Equal when they:
          have equal tickers and USD values; are partially equal and have the
          same domain; or are a match and also have the same amount value.
      - Tokens held by a single MM / Solver that are a partial match (same ticker
        and USD values) are able to be exchanged for one another during a bid
        offer. This assumes the MM is willing to take token `X_1` on domain `1` in
        exchange for token `A` and then on domain `2` exchange `X_2` for token
        `B` in the course of filling an order. For example, where `_N` represents
        Ticker `T` on domain `N` such that for any `N` ticker `T_N` and `T_M` are
        matches, a path across 2 domains may look like the following:
        - `(TOK1_1 -> TOK2_1) -> (TOK2_2 -> TOK3_2)` which details a swap of
          `TOK1` on domain 1 with `TOK3` on domain 2 via an intermediary exchange
          of `TOK2` on domains 1 and 2 by the Solver.
  - Paths and Pools
    - Pools are defined as abstractions of liquidity pools simply defining a
      pair of Tokens `A` and `B` with their respected token values and total
      volume within the "pool" instead of the amount held by a single holder
      (the total "held" by the pool).
      - Pools are defined such that there is a pair of tokens `A` and `B` where
        trading is only defined in the direction of `A` to `B` not in both
        directions. In a real pool of tokens trading would be possible in both
        directions `A -> B` and `B -> A` without limitation. This approach was
        decided due to simplicity and ease of implementation over any real world
        equivalences.
      - Pools are only mentioned in the sense that a MM / Solver may have a 
        "position" (holdings of equal value of tokens `A` and `B`) in a pool.
        These positions are defined only in terms of where a MM / Solver has a 
        pair of tokens "deposited" in a pool and are not unique or identifiable
        outside of the pair of tokens and total volume they represent. This is
        for simplicity only and does not, clearly, represent how real pools of
        tokens are defined, this allows for very simple abstractions of pools to
        be made without needing to extensively define them as they aren't the
        main focus of this project - just a part of it.
    - Paths are simply a list of pools where a MM / Solver has positions in
      the pools and is able to link those positions together such that the
      start and end token `A` and token `B` match the desired input and output
      tokens of an order.
      - When a MM / Solver proposes a bid for an order they may take the
        exchange of token `A` on domain `X` for token `B` on domain `Y` at a 
        given point if and only if the solver holds a position of token `A` in
        a pool with tokens `C` and `D` where `C` is a partial match for token
        `A`. For example: `TOK1_dom0` is desired to be swapped for `TOK3_dom1`,
        this can be filled if the solver has a position with enough liquidity
        in a pool of `TOK1_dom1` and `TOK3_dom1` or another path of positions
        that match.
  - Orders
    - Orders are defined as a pair of tokens, the `origin` and `target` tokens
      which are to be solved for. This means the `origin` token is to be swapped
      `N` times in `N` pools as defined above in a path (`1-N` in length where
      `N>=1`).
  - Users
    - Configured via the configuration file and detailed as unique agents with
      unique names paired with uniquely generated UUIDs to represent them.
    - Users are able to "post" (dispatch) orders and receive and accept "bids"
      (solutions to solve the order via trading in liquidity pools) for these
      orders. They always choose, in this PoC implementation, the first and 
      shortest valid bid. In reality the selection process would be more complex
      as is discussed later briefly.
    - Users exist solely for the purpose of posting orders in this PoC and the
      orders they post are defined in the configuration file and executed per
      user sequentially.
- Auction System
  - Pairs Orders and the Bids that come in for them concurrently from numerous
    MMs / Solvers across `N` domains such that the final token is the token
    desired to be exchanged for the original input token.
  - Concurrency
    - Concurrency is used in the form of Goroutines to allow for MMs / Solvers
      to respond to orders as they come in from Users and respond such that each
      "bid" from each Solver can be processed concurrently by the User via
      channels.
    - Currently the Order posting implementation is synchronous going through
      each user and their orders in turn and concurrently listening for bids
      until ready to accept a "winning" bid.
  - Cross / Multi - Domain Support
    - Cross-Domain support comes in the form of allowing multiple domains to be
      used to fulfil any given order such that the origin and target tokens are
      equal to the domain level. The pairing system follows a modified BFS
      searching algorithm in order to find the shortest path, for each Solver,
      to fulfil the order.
    - Bid Path Searching
      - The BFS searching algorithm used to solve an order for a specified pair
        of tokens (origin and target) specified by the User by navigating each
        Solver's positions in liquidity pools on each domain such that the
        origin and target tokens can be linked in `N` steps.
      - Once found the bid is defined and sent to the User over a specific
        channel unique to the order in question, if unable to be solved by the
        Solver then `nil` is returned to the User specifying that Solver cannot
        solve the order on their own.
      - Due to the BFS nature of the searching algorithm, the shortest possible
        path from the origin to target token is chosen and sent to the User. The
        user then in turn chooses the first, shortest of all bids it receives
        discarding any `nil` or invalid bids (defined such that at least one of
        the pools in the path was found to be invalid) before deciding.
