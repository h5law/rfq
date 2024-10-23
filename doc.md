# \[R\]equest \[F\]or \[Q\]uote

This application is a basic PoC of a cross-domain RFQ auction styled order-bid
pairing system for users and market-makers / solvers to interact with each other
efficiently in an automated manner.

The auction styled order-bid pairing system is able to find, for a specific
order, the MM / Solver and the specific path through the MM's positions in
liquidity pools across multiple domains such that the order is filled in the
shortest number of steps as is possible. However, there are many assumptions and
limitations in this implementation - which is why it is only a PoC and not a
production ready (or usable in any manner besides testing) implementation.

**This is only a PoC and is not intended to be used outside of testing at all.**

## Features, Components, Assumptions and Details

The configuration file lays out a few example cases of erroring orders, invalid
bids, competing bids, cross-domain bids and regular bids in a deterministic manner.
The errors from invalid configurations or invalid bids/orders should be caught
and either `panic`-ed or displayed to in the output respectively.

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
      `N` times in `N` pools, as defined above, in a path (`1-N` in length where
      `N>=1`) which links the `origin` token to the `target` token in this
      auction system. Thus an order can be thought of as an exchange of one
      token for another on any particular domain, token `A` and `B` . The MM /
      Solver that posts their quote to fill the order then defines the path of
      positions in liquidity pools through which the origin token will take as it
      is swapped until it reaches token `B` the order routing system is discussed
      later in more detail. Orders also contain a channel to which each solver
      is to send their bids for the order - or nil if they cannot solve it. In
      addition to some generic fields such as the order ID, user ID and solver
      ID to identify the order, its creator and solver with. As well as an entry
      and exit time where the order was created and a bid accepted along with a
      timeout period where to be filled - otherwise time'd-out orders are not filled. 
      - If no timeout is specified in order configuration, then the maximum
        possible timeout period is used which is approximately equal to 290
        years from the timestamp at the time of order creation. Essentially
        making orders with no specified timeout open forever or until all solvers
        have submitted a bid (including `nil` bids for unfilled orders).
      - In this specific system the potential for partial filling orders is very
        possible. Orders that do not get filled in their entirety are not filled
        at all in this system as it currently stands.
    - Bids are responses to open orders by solvers and are in one of two states:
      complete/full-fulfilment or `nil` no fulfilment possible - partial or multi
      solver bids are not currently possible as discussed below. Bids are
      essentially a list of the pools needed to swap in (uni-directionally) in
      order to fulfil the order by utilising the single solver's positions as
      liquidity to fill the order.
      - Bids and solvers are discussed in their parts in more detail below.
  - Users and Solvers
    - Configured via the configuration file and detailed as unique agents with
      unique names paired with uniquely generated UUIDs to represent them.
    - Users are able to "post" (dispatch) orders and receive and accept "bids"
      (solutions to solve the order via trading in liquidity pools) for these
      orders. They always choose, in this PoC implementation, the first and 
      shortest valid bid. In reality the selection process would be more complex
      as is discussed later briefly.
    - Users exist solely for the purpose of posting orders in this PoC and the
      orders they post are defined in the configuration file and executed per
      user sequentially. So user one executes all their orders sequentially then
      user two and so on, this could be replaced with a concurrent model similar
      to how solvers listen for orders very easily but would show little benefit
      in the current state of the PoC.
    - Solvers listen all concurrently for orders being posted by the users. On
      receipt of an order, via the pub-sub / consumer model of order distribution
      and receiving of orders (the auction), the solver then uses a BFS-like
      algorithm that traverses the graph of their liquidity positions to find the
      shortest possible path they can make, if any, to fully fulfil the order.
    - For simplicity solvers check they can fully fill an order at each step of
      the path they are making and thus partial fills are not possible only full
      or `nil` bids for an order can be made.
    - The decision to make orders be executed sequentially and bids concurrently
      was to somewhat mirror how users cannot execute two transactions at the
      exact same time but one after the other but many actors can see the
      transaction being executed as it happens if they are watching. This is
      reflected in the User posting and Solver listening mechanism.

- Configuration File
  - Instead of making the system interactive, in order to aid testing/debugging
    and reproducibility, a configuration file was chosen to describe the state
    of the system. Users, Solvers and Orders are defined in the config file and 
    read on startup and used to build the system according to the design laid out
    in the config. It would be very easy to replace this with an interactive
    model where the system starts and actors "join" with their own configs but
    for simplicity these are centralised into a single point of entry. This means
    testing (by using or in unit tests) are easier as a test config file can be
    supplied with deterministic states meaning test outputs can be verified.

- Auction System
  - Pairs Orders and the Bids that come in for them concurrently from numerous
    MMs / Solvers across `N` domains such that the final token is the token
    desired to be exchanged for the original input token.
  - Concurrency
    - Concurrency is used in the form of Goroutines, Channels and a Pub-Sub / 
      Consumer model in order to allow for MMs / Solvers to respond to orders
      as they come in from Users and respond such that each "bid" from each
      Solver can be processed concurrently by the User via channels unique to
      each order.
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
  - Settlement
    - No orders are settled in this system. This is a PoC of a RFQ auction style
      bidding system and thus doesn't include the settlement or finalisation of
      any orders. This means no tokens are actually "exchanged" and swapped
      through the pools of the chosen bid upon selection by the user. This is
      simply to highlight and focus on the bid selection process over anything
      else. If settlement was to be introduced then the affect would be that
      balances of tokens held / ratios of positions in liquidity pools (potentially)
      may change due to the exchange of tokens. Users balances would drop in
      the origin token and rise in the target token, solvers positions would
      react accordingly in response to the trade they executed unbalancing the pool.
      This may also open arbitrage opportunities the solvers may be able to exploit.
    - Partial Fulfilment is not possible nor is fulfilment using multiple solvers
      currently, this is again to keep the PoC simple. However to incorporate
      partial fills one would need only remove the check that at each step of
      the path in a bid satisfies the entire order's volume as well as the modify
      that the search ends when the token matches the target token. In their
      place you would put a counter that can be depleted, starting on the full
      volume and reducing every time only part of the order can be filled for each
      path in the graph and pair this with a check that the search ends on the
      target token as before but also ensure the volume is less than or equal to
      the target amount. This change allows for more solvers to submit bids for
      part of the order instead of the entire amount.
    - Multi-Solver Solutions would require more drastic changes but would enable
      greater flexibility in order resolution, by allowing combining partial fills
      or bridging solver paths opening new paths from `A` to `B` so more orders
      could be filled by numerous solvers "working together".
    - Redemption and Exchange
      - Actual settlement and the exchange of funds would work similar to how the
        system currently functions but the abstractions of tokens, domains, pools
        and others would be replaced with API interfaces to interact with the
        actual on-chain counterparts. A quote would be made and a path presented
        upon selection the user will transfer funds to the correct solver in the
        starting position who will programmatically execute a series of swaps if
        needed ideally within a single block similar to how a flash loan works.
        Unlike in this PoC the quote will cover any fees incurred by the solver
        and once the final token is available it will be transferred to the user.
        To reward the solvers they will collect fees but also by holding positions
        in liquidity pools they will potentially earn rewards in the form of fees
        from their usage doubling their collection.
      - In this PoC enabling the end user who receives the funds to differ from
        the instigating user is trivial and could be implemented with little
        changes to the current implementation. This would allow the representation
        of cross-domain swaps & transactions with varying origin and target tokens.

## Analysis

- Comparisons and Insights
  - Bid Selection
    - When a user selects a bid in this system they choose the first valid and 
      shortest (path-wise) bid that came in for an order as the winner of the
      auction. This is done via a BFS-styled algorithm where nodes in a graph
      are visited, then their neighbours until the node matches the target. This
      simple algorithm could be improved and could make use of the A* weighted
      graph traversal algorithm to incorporate fees, and find the shortest possible
      path to fill an order. Using such a system would be more efficient but also
      enable partial and multi solver bids more easily than the current mechanism
      due to the nature of the A* algorithm knowing the entire graph at once.
      However implementing the A* algorithm for this model would be far more
      complex and require much more time to be dedicated to its upkeep. Despite
      this, I believe it to be the logical next step in this systems architecture.
  - Integrations
    - In order to integrate with actual smart contracts and on-chain systems
      there are numerous parts of the code that can be removed, some replaced
      and others modified to adapt to these changes. The auction, order-bid
      pairing system should remain off-chain for efficiency (unless absolutely
      necessary that it live on-chain for some reason - in this case regular
      proofs can be posted on chain instead in a commit and reveal claim and 
      proof style lifecycle where state integrity of the auction is guarenteed
      authentic via its posted proof). Users, Solvers, Pools, Tokens/Coins,
      Domains/Chains will all be replaced with there on-chain equivalent APIs
      and libraries. Much of the boilerplate code from this model will be
      replaced with API/Library/SDK/Smart Contract Binding calls to produce the
      same but real functionality. Having designed this system with the
      abstractions it has in place of the real world entities that would take
      their place allows for an easier integration but also a better conceptual
      view of how the system operates not only in the emulated environment but
      how it would in a real one too.
  - Decision Making
    - In this system tokens across domains of the same ticker have the same USD
      value per token, this however is just an assumption made for easier
      implementation. In reality there will be slight differences in value across
      domains. Solvers must be aware of these differences, as well as current
      gas fees, potential arbitrage opportunities along the path to solve an 
      order and much more. For example, if the market maker had a complete view
      of their positions and they were able to see potential arbitrage 
      opportunities as they arose they may be able to offer the user a cheaper
      potentially longer route and claim more profits in the arbitrage than from
      user fees. By making the bid cheaper the user is more likely to accept it 
      and thus their arbitrage is more likely to occur. They must also be aware
      of any fees they will incur in gas and any other interactions they may make.
      By providing the solver with up-to-date information on the value of the
      assets being exchanged and current fees etc, the system allows for the
      solver to make better informed decisions.
    - In conjunction with the above users should also make more informed decisions
      when accepting a bid and settling an order. They should be aware of any
      fees (gas and others) they will incur as well as all possible options not
      simply the first, shortest valid bid.
