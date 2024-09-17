# Hand_in_2

This is part of the course distributed systems and security

Hand-in 2 IT 5

Exercise 4.6 and 5.1

# Peer-to-Peer Network Implementation

## Todo List

### Forming a Network

- [x] **Implement Peer structure**
  - Define the type `Peer` representing each peer in the network.

- [x] **Implement Peer.Connect(addr string, port int)**
  - [x] Allow peers to connect to the network using an IP address and port.
  - [x] If no valid peer is found, start a new network with only the peer itself.

- [x] **Print peer information**
  - [x] After connecting or creating a network, print the peer's IP address and port.

- [ ] **Set of peers**
  - [ ] Each peer should maintain a set of known peers in the network.

- [ ] **Request peer set from existing peers**
  - [ ] When a peer joins, ask the existing peer for its set of peers.

- [ ] **Join Message**
  - [ ] New peers should broadcast a "Join Message" to notify the network of their presence.
  - [ ] Existing peers should update their peer set upon receiving a "Join Message".

### Flooding a Message

#### Simple Flooding Solution

- [ ] **Implement Peer.FloodMessage(msg <some type>)**
  - [ ] Send the message to all peers in the peer set.

#### Advanced Flooding Solution

- [ ] **Advanced Message Flooding**
  - [ ] Ensure messages are sent to peers that have not yet sent the message to the sender.
  - [ ] Upon receiving a message, flood it unless the peer has already sent it.
  - [ ] When a new peer joins, send all previous messages to ensure message consistency.

### Keeping a Ledger

- [ ] **Implement local ledger**
  - Define the `Ledger` type that maintains account balances.

- [ ] **Implement Peer.FloodTransaction(tx *Transaction)**
  - [ ] Implement the flooding of transactions using the `FloodMessage` mechanism.
  - [ ] Ensure each peer executes received transactions on its local ledger.

### Demo Program

- [ ] **Implement handin.go**
  - [ ] Start a network of `n = 10` peers (or less if necessary) on the same machine.
  - [ ] Ensure peers pick different ports to avoid conflicts.
  - [ ] Ensure the program is easily runnable on the TA’s machine.

- [ ] **Send τ = 10 transactions from each peer**
  - [ ] Use 5 accounts (e.g., account1, ..., account5) for transactions.
  - [ ] All peers should send transactions related to all 5 accounts.

- [ ] **Test ledger consistency**
  - [ ] After all transactions are sent, verify that all peers hold identical ledgers.

- [ ] **Optional: Stress testing**
  - [ ] Test with larger `n` and `τ` to evaluate system limits (e.g., transactions per second).
  - [ ] Document crash or trash limits if applicable.

### Testing and Reporting

- [ ] **Testing**
  - [ ] Write automated tests for the system to verify its correctness.
  - [ ] Describe the testing procedure and results in the report.

- [ ] **Advanced Flooding: Eventual Consistency**
  - [ ] In the report, argue that the system achieves eventual consistency if no more floods are initiated.

- [ ] **Simple Flooding: Consistency Scenarios**
  - [ ] Provide a scenario where simple flooding leads to eventual consistency.
  - [ ] Provide a scenario where simple flooding does not lead to eventual consistency.

- [ ] **Transaction Rejection (Optional)**
  - [ ] Discuss how eventual consistency is affected if transactions that reduce an account's balance below 0 are rejected.
