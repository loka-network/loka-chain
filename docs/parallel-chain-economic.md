# Parallel Chain Economic Model

## 1. Model Overview

### 1.1 System Architecture

- **Main Chain**: Responsible for governance, validator management, and child chain registration
- **Child Chain**: Independent blockchain networks operated by participating validators
- **Validators**: Entities that simultaneously maintain main chain and participating child chain nodes

### 1.2 Economic Participants

- **validator_main**: Main chain validator set {v₁, v₂, ..., vₙ}
- **validator_child**: Child chain validator subset {vᵢ₁, vᵢ₂, ..., vᵢₘ} ⊆ validator_main
- **Users**: Child chain transaction initiators and fee payers

## 2. Reward Mechanism Design

### 2.1 Block Rewards

#### 2.1.1 Base Block Reward

Fixed reward per child chain block:

```
BR_base = 100 CHILD_TOKEN
```

#### 2.1.2 Inflation Adjustment Mechanism

Inflation control considering total token supply:

```
BR_adjusted(h) = BR_base × (1 - α)^⌊h/I⌋

Where:
- h: Current block height
- α: Inflation reduction rate (0.05 = 5%)
- I: Adjustment interval (1,000,000 blocks)
- ⌊⌋: Floor function
```

#### 2.1.3 Minimum Block Reward

```
BR_final(h) = max(BR_adjusted(h), BR_min)

Where:
- BR_min = 10 CHILD_TOKEN (minimum guaranteed reward)
```

### 2.2 Transaction Fees

#### 2.2.1 Gas Fee Calculation

Fee per transaction:

```
Fee_tx = Gas_used × Gas_price + Priority_fee

Where:
- Gas_used: Amount of gas consumed by the transaction
- Gas_price: Base gas price (0.001 CHILD_TOKEN/Gas)
- Priority_fee: Priority fee (optional)
```

#### 2.2.2 Total Block Fees

```
TF_block = Σᵢ₌₁ⁿ Fee_txᵢ

Where:
- n: Number of transactions in the block
- Fee_txᵢ: Fee of the i-th transaction
```

## 3. Reward Distribution Algorithm

### 3.1 Simplified Distribution Strategy

#### 3.1.1 Block Reward Distribution

Evenly distributed among all participating validators:

```
Reward_validator_block(vⱼ) = BR_final(h) / |V_child|

Where:
- |V_child|: Number of validators participating in the child chain
- vⱼ ∈ V_child: Participating validator j
```

#### 3.1.2 Fee Distribution

All fees go to the current block proposer:

```
Reward_proposer_fees = TF_block

Reward_validator_fees(vⱼ) = {
    TF_block,  if vⱼ is the current proposer
    0,         otherwise
}
```

#### 3.1.3 Total Validator Reward

```
Reward_total(vⱼ, h) = Reward_validator_block(vⱼ) + Reward_validator_fees(vⱼ)
```

### 3.2 Expected Return Calculation

#### 3.2.1 Long-term Average Return

Expected validator reward over N block cycles:

```
E[Reward_validator(vⱼ, N)] = N × (BR_final / |V_child|) + 
                              (N / |V_child|) × E[TF_block]

Where:
- E[TF_block]: Expected value of block fees
- N / |V_child|: Expected number of proposed blocks
```

#### 3.2.2 Fee Expectation Estimation

Fee expectation based on historical data:

```
E[TF_block] = μ_fee × λ_tx

Where:
- μ_fee: Average transaction fee
- λ_tx: Average number of transactions per block
```

### 3.3 Staking and Slashing Mechanism

#### 3.3.1 Minimum Staking Requirement

```
Min_stake = 1,000,000 LOKA  // Minimum stake to participate in child chain validation
```

#### 3.3.2 Slashing Mechanism

```
Slash_rate = 0.05  // Penalty rate for malicious behavior
```

### 3.4 Network Security Assurance

#### 3.4.1 Minimum Validator Requirement

```
Min_validators = 5  // Based on BFT requirements in the codebase
```

## Conclusion

This economic model provides a robust and sustainable economic foundation for parallel chain subnets. Through precise mathematical modeling and risk control, it ensures the long-term healthy development of the network.

