# go-kaspabook REST API Documentation

_Lightweight REST API with indexed Kaspa._

## Overview

This document is generated from the current `go-kaspabook` source code by AI. The API server is implemented with Fiber and exposes two groups of endpoints:

- `book` endpoints: indexed/query APIs backed by local indexed data
- `kaspad` endpoints: passthrough APIs backed by Kaspa gRPC calls

## Base Behavior

### Content Type

All successful and error responses are JSON.

### Common Response Shape

Most endpoints return:

```json
{
  "message": "successful",
  "result": {}
}
```

### Common `message` Values

- `synced`: backend and index status are considered synchronized
- `unsynced`: backend is not ready or index lag is too large
- `successful`: request completed successfully
- `failed`: request completed but logical submission failed
- `internal error`: internal indexed-data query failed
- `kaspad error`: Kaspa gRPC request failed
- `data expired`: reserved message constant in codebase
- `not reached`: reserved message constant in codebase

### Global Rules

- `GET /book/status` is always allowed
- all other endpoints return `503` with `message: "unsynced"` when Kaspa is not synced
- unmatched routes return `404`

---

## Shared Response Data Structures

## Block Object

Used by:

- `GET /book/blocks/{hash}`
- `GET /book/vspcs/daascore/{score}`
- `GET /book/vspcs/bluescore/{score}`

```json
{
  "hash": "string",
  "daaScore": "string",
  "blueScore": "string",
  "timestamp": "string",
  "acceptedIdMerkleRoot": "string",
  "isChainBlock": "string"
}
```

### Fields

| Field | Type | Meaning |
|------|------|---------|
| `hash` | string | Block hash, hex-encoded |
| `daaScore` | string | DAA score of the block |
| `blueScore` | string | Blue score of the block |
| `timestamp` | string | Block timestamp as a numeric string |
| `acceptedIdMerkleRoot` | string | Accepted transaction ID merkle root, hex-encoded |
| `isChainBlock` | string | Whether the block is treated as a chain block; currently always formatted as `"true"` |

---

## Transaction Input Object

Used inside transaction responses.

```json
{
  "prevTxId": "string",
  "prevTxIndex": "string",
  "address": "string",
  "amount": "string",
  "spk": "string",
  "spkType": "string"
}
```

### Fields

| Field | Type | Meaning |
|------|------|---------|
| `prevTxId` | string | Previous transaction ID being spent, hex-encoded |
| `prevTxIndex` | string | Output index in the previous transaction |
| `address` | string | Source Kaspa address associated with the consumed UTXO |
| `amount` | string | Input amount |
| `spk` | string | Script public key derived from the address |
| `spkType` | string | Script public key type derived from the address |

---

## Transaction Output Object

Used inside transaction responses.

```json
{
  "address": "string",
  "amount": "string",
  "spk": "string",
  "spkType": "string"
}
```

### Fields

| Field | Type | Meaning |
|------|------|---------|
| `address` | string | Destination Kaspa address |
| `amount` | string | Output amount |
| `spk` | string | Script public key derived from the address |
| `spkType` | string | Script public key type derived from the address |

---

## Transaction Object

Used by:

- `GET /book/transactions/{txid}`
- `GET /book/addresses/{address}/transactions`
- VSPC transaction lists

```json
{
  "txId": "string",
  "txHash": "string",
  "inputs": [],
  "outputs": [],
  "fee": "string",
  "blockHash": "string",
  "blockTime": "string",
  "acceptedBlock": "string",
  "acceptedDaaScore": "string",
  "acceptedBlueScore": "string",
  "acceptedTime": "string",
  "isAccepted": "string"
}
```

### Fields

| Field | Type | Meaning |
|------|------|---------|
| `txId` | string | Transaction ID, hex-encoded |
| `txHash` | string | Transaction hash, hex-encoded |
| `inputs` | array | List of transaction inputs |
| `outputs` | array | List of transaction outputs |
| `fee` | string | Calculated fee = total input amount - total output amount, if non-negative |
| `blockHash` | string | Block hash originally associated with the transaction record |
| `blockTime` | string | Block time associated with the transaction record |
| `acceptedBlock` | string | Accepting block hash, if the transaction is accepted |
| `acceptedDaaScore` | string | DAA score of the accepting block |
| `acceptedBlueScore` | string | Blue score of the accepting block |
| `acceptedTime` | string | Timestamp of the accepting block |
| `isAccepted` | string | `"true"` if accepted block data is present, otherwise `"false"` |

---

## VSPC Object

Used by:

- `GET /book/vspcs/daascore/{score}`
- `GET /book/vspcs/bluescore/{score}`

```json
{
  "block": {},
  "transactions": []
}
```

### Fields

| Field | Type | Meaning |
|------|------|---------|
| `block` | object | Block metadata for this VSPC entry |
| `transactions` | array | Transactions associated with the block in this VSPC result |

---

## Status Object

Used by:

- `GET /book/status`

```json
{
  "versionKaspad": "string",
  "daaScoreKaspad": "string",
  "statusKaspad": "string",
  "versionBook": "string",
  "daaScoreBook": "string",
  "blueScoreBook": "string",
  "scannedBook": "string",
  "gapBook": "string",
  "sizeBook": "string",
  "network": "string",
  "hysteresis": "string",
  "dtlIndex": "string",
  "totalBlock": 0,
  "totalTransaction": 0
}
```

### Fields

| Field | Type | Meaning |
|------|------|---------|
| `versionKaspad` | string | Version of the upstream kaspad node |
| `daaScoreKaspad` | string | Current kaspad DAA score |
| `statusKaspad` | string | Current kaspad sync/runtime status |
| `versionBook` | string | Version of the local indexed service/book |
| `daaScoreBook` | string | Indexed book DAA score |
| `blueScoreBook` | string | Indexed book blue score |
| `scannedBook` | string | Internal scanned progress marker for the book index |
| `gapBook` | string | Gap between kaspad and indexed book progress |
| `sizeBook` | string | Indexed database size or size-related runtime value |
| `network` | string | Network name, such as mainnet/testnet/devnet/simnet |
| `hysteresis` | string | Configured hysteresis threshold |
| `dtlIndex` | string | Indexed data lifetime configuration |
| `totalBlock` | integer | Total block count; current `/book/status` handler forces this to `0` before response |
| `totalTransaction` | integer | Total transaction count; current `/book/status` handler forces this to `0` before response |

---

## Balance Entry Object

Used by:

- `GET /kaspad/addresses/{address}/balance`

```json
{
  "address": "string",
  "balance": 0,
  "error": {}
}
```

### Fields

| Field | Type | Meaning |
|------|------|---------|
| `address` | string | Queried Kaspa address |
| `balance` | uint64 | Current balance for the address |
| `error` | object/null | Per-entry RPC error from kaspad, if any |

---

## UTXO Entry Object

Used by:

- `GET /kaspad/addresses/{address}/utxos`

```json
{
  "address": "string",
  "outpoint": {},
  "utxoEntry": {}
}
```

### Fields

| Field | Type | Meaning |
|------|------|---------|
| `address` | string | Kaspa address owning this UTXO |
| `outpoint` | object | Reference to the transaction output being spent later |
| `utxoEntry` | object | Details of the UTXO itself |

### `outpoint` Fields

```json
{
  "transactionId": "string",
  "index": 0
}
```

| Field | Type | Meaning |
|------|------|---------|
| `transactionId` | string | Transaction ID containing the output |
| `index` | uint32 | Output index in the transaction |

### `utxoEntry` Fields

```json
{
  "amount": 0,
  "scriptPublicKey": {},
  "blockDaaScore": 0,
  "isCoinbase": false,
  "verboseData": {}
}
```

| Field | Type | Meaning |
|------|------|---------|
| `amount` | uint64 | UTXO amount |
| `scriptPublicKey` | object | Raw script public key data |
| `blockDaaScore` | uint64 | DAA score of the block containing this UTXO |
| `isCoinbase` | bool | Whether this UTXO originated from a coinbase transaction |
| `verboseData` | object/null | Human-readable script metadata |

### `scriptPublicKey` Fields

```json
{
  "version": 0,
  "scriptPublicKey": "string"
}
```

| Field | Type | Meaning |
|------|------|---------|
| `version` | uint32 | Script public key version |
| `scriptPublicKey` | string | Script public key bytes/content |

### `verboseData` Fields

```json
{
  "scriptPublicKeyType": "string",
  "scriptPublicKeyAddress": "string"
}
```

| Field | Type | Meaning |
|------|------|---------|
| `scriptPublicKeyType` | string | Script type |
| `scriptPublicKeyAddress` | string | Decoded address for the script |

---

## Submit Transaction Response Object

Used by:

- `POST /kaspad/transactions`

```json
{
  "transactionId": "string",
  "error": {}
}
```

### Fields

| Field | Type | Meaning |
|------|------|---------|
| `transactionId` | string | Submitted transaction ID |
| `error` | object/null | RPC-level logical error returned by kaspad |

---

## Submit Transaction Replacement Response Object

Used by:

- `POST /kaspad/transactions/rbf`

```json
{
  "transactionId": "string",
  "replacedTransaction": {},
  "error": {}
}
```

### Fields

| Field | Type | Meaning |
|------|------|---------|
| `transactionId` | string | Submitted replacement transaction ID |
| `replacedTransaction` | object/null | Previous transaction replaced in the mempool |
| `error` | object/null | RPC-level logical error returned by kaspad |

---

## Endpoints

## GET /book/status

Returns synchronization and runtime status.

### Response

```json
{
  "message": "synced",
  "result": {
    "versionKaspad": "string",
    "daaScoreKaspad": "string",
    "statusKaspad": "string",
    "versionBook": "string",
    "daaScoreBook": "string",
    "blueScoreBook": "string",
    "scannedBook": "string",
    "gapBook": "string",
    "sizeBook": "string",
    "network": "string",
    "hysteresis": "string",
    "dtlIndex": "string",
    "totalBlock": 0,
    "totalTransaction": 0
  }
}
```

### Field Meaning

- `message`: overall sync state returned by the API
- `result`: status object described above

---

## GET /book/blocks/{hash}

Returns a single indexed block.

### Response

```json
{
  "message": "successful",
  "result": {
    "hash": "string",
    "daaScore": "string",
    "blueScore": "string",
    "timestamp": "string",
    "acceptedIdMerkleRoot": "string",
    "isChainBlock": "string"
  }
}
```

### Field Meaning

- `message`: request result status
- `result.hash`: block hash
- `result.daaScore`: DAA score
- `result.blueScore`: blue score
- `result.timestamp`: block timestamp
- `result.acceptedIdMerkleRoot`: accepted transaction ID merkle root
- `result.isChainBlock`: chain block indicator

---

## GET /book/transactions/{txid}

Returns a single indexed transaction.

### Response

```json
{
  "message": "successful",
  "result": {
    "txId": "string",
    "txHash": "string",
    "inputs": [],
    "outputs": [],
    "fee": "string",
    "blockHash": "string",
    "blockTime": "string",
    "acceptedBlock": "string",
    "acceptedDaaScore": "string",
    "acceptedBlueScore": "string",
    "acceptedTime": "string",
    "isAccepted": "string"
  }
}
```

### Field Meaning

- `message`: request result status
- `result.txId`: transaction ID
- `result.txHash`: transaction hash
- `result.inputs`: consumed previous outputs
- `result.outputs`: newly created outputs
- `result.fee`: transaction fee
- `result.blockHash`: source block hash associated with transaction record
- `result.blockTime`: source block time associated with transaction record
- `result.acceptedBlock`: block that accepted this transaction
- `result.acceptedDaaScore`: accepting block DAA score
- `result.acceptedBlueScore`: accepting block blue score
- `result.acceptedTime`: accepting block timestamp
- `result.isAccepted`: whether accepted block metadata is present

---

## GET /book/vspcs/daascore/{score}

Returns VSPC entries ordered by DAA score.

### Response

```json
{
  "message": "successful",
  "result": [
    {
      "block": {
        "hash": "string",
        "daaScore": "string",
        "blueScore": "string",
        "timestamp": "string",
        "acceptedIdMerkleRoot": "string",
        "isChainBlock": "string"
      },
      "transactions": [
        {
          "txId": "string",
          "txHash": "string",
          "inputs": [],
          "outputs": [],
          "fee": "string",
          "blockHash": "string",
          "blockTime": "string",
          "acceptedBlock": "string",
          "acceptedDaaScore": "string",
          "acceptedBlueScore": "string",
          "acceptedTime": "string",
          "isAccepted": "string"
        }
      ]
    }
  ]
}
```

### Field Meaning

- `message`: request result status
- `result[].block`: block metadata for the VSPC item
- `result[].transactions`: transactions associated with that VSPC block

---

## GET /book/vspcs/bluescore/{score}

Returns VSPC entries ordered by blue score.

### Response Structure

Same as `GET /book/vspcs/daascore/{score}`.

### Field Meaning

Same as above.

---

## GET /book/addresses/{address}/transactions

Returns transactions associated with a specific address.

### Response

```json
{
  "message": "successful",
  "result": [
    {
      "txId": "string",
      "txHash": "string",
      "inputs": [
        {
          "prevTxId": "string",
          "prevTxIndex": "string",
          "address": "string",
          "amount": "string",
          "spk": "string",
          "spkType": "string"
        }
      ],
      "outputs": [
        {
          "address": "string",
          "amount": "string",
          "spk": "string",
          "spkType": "string"
        }
      ],
      "fee": "string",
      "blockHash": "string",
      "blockTime": "string",
      "acceptedBlock": "string",
      "acceptedDaaScore": "string",
      "acceptedBlueScore": "string",
      "acceptedTime": "string",
      "isAccepted": "string"
    }
  ]
}
```

### Field Meaning

- `message`: request result status
- `result[]`: formatted transaction objects for the address history

#### `inputs[]`

- `prevTxId`: previous transaction ID referenced by this input
- `prevTxIndex`: output index referenced by this input
- `address`: address associated with the spent input
- `amount`: spent amount
- `spk`: derived script public key
- `spkType`: derived script type

#### `outputs[]`

- `address`: recipient address
- `amount`: output amount
- `spk`: derived script public key
- `spkType`: derived script type

---

## GET /kaspad/addresses/{address}/balance

Returns the balance for one address.

### Response

```json
{
  "message": "successful",
  "result": {
    "address": "string",
    "balance": 0,
    "error": {}
  }
}
```

### Field Meaning

- `message`: request result status
- `result.address`: queried address
- `result.balance`: current balance
- `result.error`: per-entry backend error, if present

---

## GET /kaspad/addresses/{address}/utxos

Returns all UTXOs for one address.

### Response

```json
{
  "message": "successful",
  "result": [
    {
      "address": "string",
      "outpoint": {
        "transactionId": "string",
        "index": 0
      },
      "utxoEntry": {
        "amount": 0,
        "scriptPublicKey": {
          "version": 0,
          "scriptPublicKey": "string"
        },
        "blockDaaScore": 0,
        "isCoinbase": false,
        "verboseData": {
          "scriptPublicKeyType": "string",
          "scriptPublicKeyAddress": "string"
        }
      }
    }
  ]
}
```

### Field Meaning

- `message`: request result status
- `result[]`: UTXO records owned by the address
- `result[].address`: owner address
- `result[].outpoint.transactionId`: transaction containing the UTXO
- `result[].outpoint.index`: output index
- `result[].utxoEntry.amount`: UTXO amount
- `result[].utxoEntry.scriptPublicKey.version`: script version
- `result[].utxoEntry.scriptPublicKey.scriptPublicKey`: raw script public key
- `result[].utxoEntry.blockDaaScore`: block DAA score
- `result[].utxoEntry.isCoinbase`: whether coinbase-derived
- `result[].utxoEntry.verboseData.scriptPublicKeyType`: script type
- `result[].utxoEntry.verboseData.scriptPublicKeyAddress`: decoded script address

---

## POST /kaspad/transactions

Submits a new transaction.

### Response

```json
{
  "message": "successful",
  "result": {
    "transactionId": "string",
    "error": {}
  }
}
```

### Field Meaning

- `message`:  
  - `successful` if submission succeeded and response error is empty  
  - `failed` if kaspad returned a logical transaction error
- `result.transactionId`: submitted transaction ID
- `result.error`: backend logical error, if any

---

## POST /kaspad/transactions/rbf

Submits a replacement transaction using RBF.

### Response

```json
{
  "message": "successful",
  "result": {
    "transactionId": "string",
    "replacedTransaction": {},
    "error": {}
  }
}
```

### Field Meaning

- `message`:  
  - `successful` if replacement succeeded and response error is empty  
  - `failed` if kaspad returned a logical transaction error
- `result.transactionId`: replacement transaction ID
- `result.replacedTransaction`: previous mempool transaction replaced by the new one
- `result.error`: backend logical error, if any

---

## Registered Routes Summary

| Method | Path |
|--------|------|
| GET | `/book/status` |
| GET | `/book/blocks/:hash` |
| GET | `/book/transactions/:txid` |
| GET | `/book/vspcs/daascore/:score` |
| GET | `/book/vspcs/bluescore/:score` |
| GET | `/book/addresses/:address/transactions` |
| GET | `/kaspad/addresses/:address/balance` |
| GET | `/kaspad/addresses/:address/utxos` |
| POST | `/kaspad/transactions` |
| POST | `/kaspad/transactions/rbf` |

---

## Source Basis

This documentation is based on the current implementation in:

- `api/init.go`
- `api/routeBookStatus.go`
- `api/routeBookAddress.go`
- `api/routeBookVspc.go`
- `api/routeKaspadAddress.go`
- `api/routeKaspadTransaction.go`
- `api/format.go`
- `database/runtime.go`
- `proto/protowire/rpc.proto`