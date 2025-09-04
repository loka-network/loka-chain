# Changelog

## [Unreleased]

- The project is under continuous development

## [2025-09-03]

### Documentation

- chore: add macos sdk minimum require

## [2025-08-29]

### Documentation

- feat: more efficient to create ante handlers
- Update Parallel Chain Economic Model
- add Economic Model for Parallel Sub-chain
- Merge remote changes from origin/main
- resolve conflicts

## [2025-08-27]

### Documentation and Scripts

- chore: update the script
- add docs

## [2025-08-25]

### Improvements

- Upgrade RocksDB and enable object store support
  - Updates RocksDB version to 10.2.1
  - Enables object store functionality across the codebase
  - Refactoring store key initialization into a separate function
  - Updating snapshot store handling to use the new object store interface

### State Machine and Store Breaking

- Adding new test targets for memiavl, store and versiondb components
- Merge branch 'loka-main' into loka
- Update README.md

## [2025-08-21]

### Scripts Update

- chore: add rocksdb build script

## [2025-08-20]

### Improvements

- Improves gas handling and block execution efficiency
  - Removes block gas meter and moves gas-wanted checks to proposal phase
  - Adds error logging for transactions exceeding block gas limit
  - Caches EVM block configuration parameters to reduce repeated allocations
  - Optimizes gas tracking by removing cumulative gas calculations
  - Extends TPS counter report period to 10 seconds for better metrics

## [2025-08-13]

### Refactor

- refactor: transfermation to the loka network and code-level renaming

## [2025-08-06]

### Improvements & State Machine Breaking

- Enhance EVM keeper with object store integration
  - Introduced object store for EVM parameters and gas tracking
  - Updated EVMConfig to retrieve and store parameters in the object store
  - Refactored gas usage tracking to utilize the object store instead of transient store
  - Added methods for managing bloom filters and gas used in transactions
  - Implemented a new nativeChange struct for tracking state changes during native actions
  - Updated interfaces to reflect new methods for balance and state management
- Improve type safety and error handling in int.go
  - Added new safe conversion functions for various integer types
  - Introduced MaxInt256 constant for 256-bit integer limits
  - Enhanced existing functions to ensure safe conversions and prevent overflows
- Enable async check tx in server configuration
  - Added configuration option for enabling async transaction checks in the server
- Removed or commented out unused code sections in various files for better clarity
- Add monitor of the TPS shell script and project document
  - show_tps.sh: the monitor of the TPS shell script

## [2025-07-30]

### State Machine Breaking

- Refactor EVM keeper to enhance fee management and balance operations
  - Updated DeductTxCostsFromUserBalance to use a new DeductFees function for better fee deduction handling
  - Introduced virtual coin transfer methods in the bank keeper for improved transaction handling
  - Added methods for balance management (Transfer, AddBalance, SubBalance, SetBalance) in the EVM keeper
  - Enhanced state management in StateDB to support native actions and event emission
  - Implemented a new nativeChange struct for tracking state changes during native actions
  - Updated interfaces to reflect new methods for balance and state management

## [2025-07-30]

### Improvements

- Refactor app to remove simapp dependency and update GenesisState type
  - Removed import of simapp from multiple files
  - Updated EthSetup and related functions to use the new GenesisState type
  - Adjusted test setup functions to align with the new GenesisState structure
  - Modified local_node.sh to change the HOMEDIR path for configuration
  - Updated go.mod to reflect changes in dependencies and versions
  - Commented out fee denomination registration in config.go

## [2025-07-15]

### Improvements

- Add go-block-stm dependency and update collections import
  - Moves cosmossdk.io/collections from indirect to direct dependency
  - Adds github.com/crypto-org-chain/go-block-stm for enhanced block processing capabilities

## [2025-01-27]

### Initial Codebase

- Initial commit with project files

## [2023-08-17]

### Initial

- Initial commit
