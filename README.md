# Static Proof-of-Stake Blockchain  

## Overview  

This repository contains the implementation of **Exercise 16.2**, the final step in a series of exercises focused on building a **total-order broadcast system** using a **Nakamoto-style blockchain**. The project extends the distributed ledger from **Exercise 6.16**, adding **tree-based total-order broadcast** and **proof-of-stake** consensus to ensure transaction ordering and prevent overdrafts.  

## Features  

- **Tree-Based Blockchain**: Implements a structured block tree ensuring a consistent transaction order.  
- **Proof-of-Stake Lottery**: Block creation is determined by a lottery where tickets correspond to the initial balance in the genesis block.  
- **Transaction Validation**: Blocks only accept correctly signed transactions that do not cause negative balances.  
- **Fixed Block Generation Time**: Blocks are created approximately every **10 seconds** to maintain network stability.  
- **Transaction Fees**: The receiver gets **1 AU** less than sent, acting as a transaction fee.  
- **Block Rewards**: Block creators receive **10 AU** plus **1 AU per transaction** as an incentive.  

## System Rules  

1. **Currency**: Transactions are in **AU**, an integral unit.  
2. **Genesis Block**:  
   - Contains **10 special accounts**, each starting with **1,000,000 AU**.  
   - Hardcoded **initial seed** for the proof-of-stake lottery.  
3. **Transactions**:  
   - Minimum transaction amount is **1 AU**.  
   - Blocks validate transactions to prevent negative balances.  
4. **Block Structure**:  
   - Blocks hold up to `BlockSize` transactions.  
   - Blocks can be sent before reaching full capacity.  
5. **Proof-of-Stake Lottery**:  
   - Peers need a **positive balance** to participate.  
   - Ticket count is based on the genesis block balance.  
   - Only the **10 initial accounts** can participate.  
6. **Security Considerations**:  
   - The same key pair is used for transactions and lottery participation.  
   - In real-world systems, separate keys would enhance security.  
7. **Network Setup**:  
   - The system runs with **10 peers**, each controlling **one initial account**.  
   - Hardness is set to generate a **new block every ~10 seconds**.  
