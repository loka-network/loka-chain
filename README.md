<!--
parent:
  order: false
-->

<div align="center">
  <h1> Loka Chain </h1>
</div>

The Loka blockchain is a fully EVM-compatible chain utilizing NeoBFT consensus and Block-STM parallel execution. This combination delivers fast finality, deterministic outcomes, and high throughput while preserving full compatibility with Ethereum tooling and smart contracts.

## LOKA Network - High-Performance Web3 Financial Infrastructure powered by Parachains

🚀 **Overview**

LOKA Network is a cutting-edge Web3 financial infrastructure designed for high-performance, scalability, and enterprise-grade reliability. LOKA leverages a modular parachain architecture with a shared consensus mechanism to enable parallel execution, rapid settlement, and flexible customization for diverse financial applications.

We aim to resolve the "three fractures" in enterprise on-chain business: slow settlement, isolated execution, and fragmented liquidity, by offering a unified, high-throughput, and secure blockchain engine.

## ✨ Key Features & Advantages

### High-Performance Parachains:
- **Sequencer**: Up to 300,000 TPS for core operations
- **Horizontal Scalability**: Each parachain capable of 100,000 TPS with sub-second finality
- **Parallel Execution (Block-STM)**: Optimistic concurrency with conflict detection for safe, parallel transaction execution and deterministic commits
- **Multi-VM Support**: Fully EVM-compatible (Solidity, Ethereum JSON-RPC, popular tooling) with optional WASM support

### Modular & Customizable:
- Deploy dedicated L2s (parachains) tailored for specific use cases (e.g., payments, high-frequency trading, gaming, compliance)
- Flexible Gas models and governance rules per sub-chain

### Consensus & Execution (NeoBFT + Block-STM):
- **NeoBFT**: Byzantine Fault Tolerant consensus providing rapid, deterministic finality and high throughput
- **Block-STM Parallel Execution**: Optimistic concurrency control with fine-grained conflict detection for parallel transaction processing
- Optional TEE-secured validators and coordinators (Intel SGX / AWS Nitro Enclaves) for verifiable execution and enhanced security, where required

### EVM Compatibility:
- Run Solidity smart contracts without modification
- Standard Ethereum JSON-RPC endpoints for wallets and infra
- Works out-of-the-box with MetaMask, Hardhat, Foundry, and common Ethereum tooling

### Enterprise-Grade Capabilities:
- **LOKA Stable**: An optional, modular feature for compliant enterprise stablecoin issuance with real-time auditability
- **LOKA DEX**: High-performance trading engine with 1ms matching latency, on-chain order book, and unified liquidity
- **Account Abstraction (AA) & Gas Abstraction**: Seamless user experience with sponsored transactions and multi-asset gas payments
- **Built-in Compliance**: KYC whitelisting, audit logs

## 🏗️ Architecture Highlights

LOKA operates on a three-layer model:

1. **Chain-S (Main Settlement Chain)**: The core hub for state anchoring, cross-chain communication (IBC-style), and optional core asset settlement
2. **Parachain Execution Layer**: Independent, high-performance sub-chains (L2s) for specific applications, running in parallel
3. **Consensus and Execution Layer**: NeoBFT consensus with Block-STM parallel execution ensures secure, consistent, and high-throughput operation across all chains; optional TEE support for verifiable nodes

## Installation

For prerequisites and detailed build instructions please read the Installation instructions. Once the dependencies are installed, run:

```bash
make install
```

Or check out the latest release.

## Deployment

**Important**: Before deploying, ensure that NeoBFT consensus and the Block-STM execution engine are properly configured. See `loka-consensus` for details.

### Local Deployment

To deploy locally, use the `local_node.sh` script. This script will set up a local environment for running the LOKA Chain node.

```bash
./local_node.sh
```

### Remote Deployment

For remote deployment, ensure that `lokad` is available in the PATH on each machine. It is also recommended to set up SSH keys on the remote machines for secure and passwordless access.

**init_validators.sh**: This script initializes the validators required for the LOKA Chain. You need to provide the remote IPs for the 4 validators in the network as parameters.

```bash
./init_validators.sh <remote_ip1> <remote_ip2> <remote_ip3> <remote_ip4>
```

**start_node_archive.sh**: This script starts the node in archive mode.

```bash
./start_node_archive.sh
```

These scripts will help you set up and run the LOKA Chain on a remote server.

## Community

The following chat channels and forums are a great spot to ask questions about LOKA Chain:

- [Open an Issue](https://github.com/loka-network/loka-chain/issues)
- [LOKA Protocol](https://www.lokachain.org/)
- [Follow us on Twitter](https://x.com/lokachain)

## Contributing

We welcome all contributions! There are many ways to contribute to the project, including but not limited to:

- Cloning code repo and opening a PR
- Submitting feature requests or bugs
- Improving our product or contribution documentation
- Contributing use cases to a feature request

For additional instructions, standards and style guides, please refer to the [Contributing document](CONTRIBUTING.md).

## Support

If you have any questions or need support, please:
1. Check our [documentation](https://www.lokachain.org/docs)
2. Open an issue on GitHub
3. Join our community discussions

---

**Built with ❤️ by the LOKA Network Team**