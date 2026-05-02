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

Most endpoints return the following structure:

```json
{
  "message": "successful",
  "result": {}
}
```

Typical `message` values found in the code:

- `synced`
- `unsynced`
- `successful`
- `failed`
- `internal error`
- `kaspad error`
- `data expired`
- `not reached`

### Global Middleware Rules

The server applies the following behavior globally:

- CORS:
  - `Access-Control-Allow-Origin: *`
  - `Access-Control-Allow-Methods: GET`
  - `Access-Control-Allow-Headers: Origin, X-Requested-With, Content-Type, Accept`
- Request timeout is controlled by server config
- Request rate/concurrency limiting is enabled
- Panic recovery is enabled
- Any route other than `GET /book/status` will return `503` with message `unsynced` if the upstream Kaspa node is not synced

### Not Found

Any unmatched route returns:

- HTTP `404`

---

## Endpoints

## GET /book/status

Returns the current API/index synchronization status.

### Description

This is the only route that remains accessible even when the Kaspa backend is unsynced.

### Response

- `200 OK` when status is available
- `503 Service Unavailable` when Kaspad is not synced

### Example Response

```json
{
  "message": "synced",
  "result": {
    "versionKaspad": "x.y.z",
    "daaScoreKaspad": "123456",
    "statusKaspad": "synced",
    "versionBook": "x.y.z",
    "daaScoreBook": "123450",
    "blueScoreBook": "123440",
    "scannedBook": "123450",
    "gapBook": "6",
    "sizeBook": "123456789",
    "network": "mainnet",
    "hysteresis": "100",
    "dtlIndex": "86400000"
  }
}
```

### Result Field Definitions

| Field | Type | Meaning |
|------|------|---------|
| `versionKaspad` | string | Version of the connected Kaspad node |
| `daaScoreKaspad` | string | Current DAA score reported by Kaspad |
| `statusKaspad` | string | Sync state of Kaspad, such as `synced` |
| `versionBook` | string | Version of the local indexed book data |
| `daaScoreBook` | string | Current indexed DAA score in the local book |
| `blueScoreBook` | string | Current indexed blue score in the local book |
| `scannedBook` | string | Indexed/scanned progress indicator for the book |
| `gapBook` | string | Gap between Kaspad state and indexed book state |
| `sizeBook` | string | Size of indexed book data |
| `network` | string | Network name, such as `mainnet`, `testnet`, etc. |
| `hysteresis` | string | Configured hysteresis threshold |
| `dtlIndex` | string | Configured index data lifetime value |

### Notes

- If `GapBook > Hysteresis + 300`, the endpoint returns `"message": "unsynced"` even if `StatusKaspad` is `"synced"`.

---

## GET /book/blocks/{hash}

Returns indexed block details by block hash.

### Path Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `hash` | string | Yes | Block hash |

### Success Response

```json
{
  "message": "successful",
  "result": {
    "hash": "block hash",
    "daaScore": "123456",
    "blueScore": "123450",
    "timestamp": "1710000000",
    "acceptedIdMerkleRoot": "merkle root hex",
    "isChainBlock": "true"
  }
}
```

### Result Field Definitions

| Field | Type | Meaning |
|------|------|---------|
| `hash` | string | Block hash in hex format |
| `daaScore` | string | DAA score of the block |
| `blueScore` | string | Blue score of the block |
| `timestamp` | string | Block timestamp |
| `acceptedIdMerkleRoot` | string | Accepted ID merkle root in hex format |
| `isChainBlock` | string | Whether the block is treated as a chain block; current formatter sets this to `"true"` |

### Empty Result Response

If no block is found:

```json
{
  "message": "successful",
  "result": null
}
```

### Error Responses

#### Invalid Hash

- HTTP `400`

```json
{
  "message": "hash invalid",
  "result": null
}
```

#### Internal Error

- HTTP `503`

```json
{
  "message": "internal error",
  "result": null
}
```

---

## GET /book/transactions/{txid}

Returns indexed transaction details by transaction ID.

### Path Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `txid` | string | Yes | Transaction hash |

### Success Response

```json
{
  "message": "successful",
  "result": {
    "txId": "transaction id",
    "txHash": "transaction hash",
    "inputs": [],
    "outputs": [],
    "fee": "0",
    "blockHash": "block hash",
    "blockTime": "1710000000",
    "acceptedBlock": "accepting block hash",
    "acceptedDaaScore": "123456",
    "acceptedBlueScore": "123450",
    "acceptedTime": "1710000001",
    "isAccepted": "true"
  }
}
```

### Result Field Definitions

| Field | Type | Meaning |
|------|------|---------|
| `txId` | string | Transaction ID in hex |
| `txHash` | string | Transaction hash in hex |
| `inputs` | array | Transaction input list |
| `outputs` | array | Transaction output list |
| `fee` | string | Computed fee as `sum(inputs) - sum(outputs)` when available |
| `blockHash` | string | Original block hash recorded on the transaction |
| `blockTime` | string | Original block timestamp recorded on the transaction |
| `acceptedBlock` | string | Hash of the block that accepted this transaction |
| `acceptedDaaScore` | string | DAA score of the accepting block |
| `acceptedBlueScore` | string | Blue score of the accepting block |
| `acceptedTime` | string | Timestamp of the accepting block |
| `isAccepted` | string | `"true"` if an accepting block is known, otherwise `"false"` |

### Input Object Field Definitions

Each item in `inputs` contains:

| Field | Type | Meaning |
|------|------|---------|
| `prevTxId` | string | Previous transaction ID referenced by this input |
| `prevTxIndex` | string | Output index referenced from the previous transaction |
| `address` | string | Address inferred/stored for the input |
| `amount` | string | Input amount |
| `spk` | string | Script public key derived from the address |
| `spkType` | string | Script public key type derived from the address |

### Output Object Field Definitions

Each item in `outputs` contains:

| Field | Type | Meaning |
|------|------|---------|
| `address` | string | Recipient/output address |
| `amount` | string | Output amount |
| `spk` | string | Script public key derived from the output address |
| `spkType` | string | Script public key type derived from the output address |

### Empty Result Response

If the transaction is not found:

```json
{
  "message": "successful",
  "result": null
}
```

### Error Responses

#### Invalid Transaction ID

- HTTP `400`

```json
{
  "message": "txId invalid",
  "result": null
}
```

#### Internal Error

- HTTP `503`

```json
{
  "message": "internal error",
  "result": null
}
```

---

## GET /book/vspcs/daascore/{score}

Returns indexed VSPC records using DAA score ordering.

### Path Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `score` | uint64 | Yes | Starting DAA score |

### Query Parameters

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `count` | integer | No | `10` | Number of records to return |
| `prev` | string | No | `""` | Use `"1"` to paginate backward |

### Success Response

```json
{
  "message": "successful",
  "result": [
    {
      "block": {
        "hash": "block hash",
        "daaScore": "123456",
        "blueScore": "123450",
        "timestamp": "1710000000",
        "acceptedIdMerkleRoot": "hex",
        "isChainBlock": "true"
      },
      "transactions": [
        {
          "txId": "transaction id",
          "txHash": "transaction hash",
          "inputs": [],
          "outputs": [],
          "fee": "0",
          "blockHash": "block hash",
          "blockTime": "1710000000",
          "acceptedBlock": "accepting block hash",
          "acceptedDaaScore": "123456",
          "acceptedBlueScore": "123450",
          "acceptedTime": "1710000000",
          "isAccepted": "true"
        }
      ]
    }
  ]
}
```

### Result Item Field Definitions

Each item in `result` contains:

| Field | Type | Meaning |
|------|------|---------|
| `block` | object | The VSPC-related block information |
| `transactions` | array | Transactions associated with that VSPC block |

### `block` Field Definitions

| Field | Type | Meaning |
|------|------|---------|
| `hash` | string | Block hash in hex |
| `daaScore` | string | Block DAA score |
| `blueScore` | string | Block blue score |
| `timestamp` | string | Block timestamp |
| `acceptedIdMerkleRoot` | string | Accepted ID merkle root in hex |
| `isChainBlock` | string | Whether the block is treated as a chain block |

### `transactions` Item Field Definitions

The structure is the same as in `GET /book/transactions/{txid}`:

- `txId`
- `txHash`
- `inputs`
- `outputs`
- `fee`
- `blockHash`
- `blockTime`
- `acceptedBlock`
- `acceptedDaaScore`
- `acceptedBlueScore`
- `acceptedTime`
- `isAccepted`

See that section for detailed field meanings.

### Empty Result Response

```json
{
  "message": "successful",
  "result": null
}
```

### Error Responses

#### Invalid Score

- HTTP `400`

```json
{
  "message": "score invalid",
  "result": null
}
```

#### Internal Error

- HTTP `503`

```json
{
  "message": "internal error",
  "result": null
}
```

---

## GET /book/vspcs/bluescore/{score}

Returns indexed VSPC records using blue score ordering.

### Path Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `score` | uint64 | Yes | Starting blue score |

### Query Parameters

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `count` | integer | No | `10` | Number of records to return |
| `prev` | string | No | `""` | Use `"1"` to paginate backward |

### Success Response

```json
{
  "message": "successful",
  "result": [
    {
      "block": {
        "hash": "block hash",
        "daaScore": "123456",
        "blueScore": "123450",
        "timestamp": "1710000000",
        "acceptedIdMerkleRoot": "hex",
        "isChainBlock": "true"
      },
      "transactions": [
        {
          "txId": "transaction id",
          "txHash": "transaction hash",
          "inputs": [],
          "outputs": [],
          "fee": "0",
          "blockHash": "block hash",
          "blockTime": "1710000000",
          "acceptedBlock": "accepting block hash",
          "acceptedDaaScore": "123456",
          "acceptedBlueScore": "123450",
          "acceptedTime": "1710000000",
          "isAccepted": "true"
        }
      ]
    }
  ]
}
```

### Result Item Field Definitions

Each item in `result` contains:

| Field | Type | Meaning |
|------|------|---------|
| `block` | object | The VSPC-related block information |
| `transactions` | array | Transactions associated with that VSPC block |

### `block` Field Definitions

| Field | Type | Meaning |
|------|------|---------|
| `hash` | string | Block hash in hex |
| `daaScore` | string | Block DAA score |
| `blueScore` | string | Block blue score |
| `timestamp` | string | Block timestamp |
| `acceptedIdMerkleRoot` | string | Accepted ID merkle root in hex |
| `isChainBlock` | string | Whether the block is treated as a chain block |

### `transactions` Item Field Definitions

The structure is the same as in `GET /book/transactions/{txid}`.

### Empty Result Response

```json
{
  "message": "successful",
  "result": null
}
```

### Error Responses

#### Invalid Score

- HTTP `400`

```json
{
  "message": "score invalid",
  "result": null
}
```

#### Internal Error

- HTTP `503`

```json
{
  "message": "internal error",
  "result": null
}
```

---

## GET /book/addresses/{address}/transactions

Returns indexed transactions associated with an address.

### Path Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `address` | string | Yes | Kaspa address |

### Query Parameters

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `count` | integer | No | `50` | Number of records to return |
| `prev` | string | No | `""` | Use `"1"` to paginate backward |
| `daascore` | uint64 | No | `0` or `18446744073709551614` | Range start when using DAA score |
| `bluescore` | uint64 | No | `0` or `18446744073709551614` | Range start when using blue score |

### Range Selection Rules

The route defaults to `daascore` mode.

- If `daascore` is absent and `bluescore` is present, the route uses blue score ordering
- If `prev=1`, the default range start becomes `18446744073709551614`
- Otherwise, the default range start is `0`

### Success Response

```json
{
  "message": "successful",
  "result": [
    {
      "txId": "transaction id",
      "txHash": "transaction hash",
      "inputs": [],
      "outputs": [],
      "fee": "0",
      "blockHash": "block hash",
      "blockTime": "1710000000",
      "acceptedBlock": "accepting block hash",
      "acceptedDaaScore": "123456",
      "acceptedBlueScore": "123450",
      "acceptedTime": "1710000000",
      "isAccepted": "true"
    }
  ]
}
```

### Result Item Field Definitions

Each transaction object contains:

| Field | Type | Meaning |
|------|------|---------|
| `txId` | string | Transaction ID in hex |
| `txHash` | string | Transaction hash in hex |
| `inputs` | array | Transaction input list |
| `outputs` | array | Transaction output list |
| `fee` | string | Computed fee |
| `blockHash` | string | Original block hash recorded on the transaction |
| `blockTime` | string | Original block timestamp recorded on the transaction |
| `acceptedBlock` | string | Hash of the accepting block |
| `acceptedDaaScore` | string | DAA score of the accepting block |
| `acceptedBlueScore` | string | Blue score of the accepting block |
| `acceptedTime` | string | Timestamp of the accepting block |
| `isAccepted` | string | Whether the transaction is accepted |

### Input Object Field Definitions

| Field | Type | Meaning |
|------|------|---------|
| `prevTxId` | string | Previous transaction ID |
| `prevTxIndex` | string | Previous output index |
| `address` | string | Input address |
| `amount` | string | Input amount |
| `spk` | string | Derived script public key |
| `spkType` | string | Derived script public key type |

### Output Object Field Definitions

| Field | Type | Meaning |
|------|------|---------|
| `address` | string | Output address |
| `amount` | string | Output amount |
| `spk` | string | Derived script public key |
| `spkType` | string | Derived script public key type |

### Empty Result Response

```json
{
  "message": "successful",
  "result": null
}
```

### Error Responses

#### Invalid Address

- HTTP `400`

```json
{
  "message": "address invalid",
  "result": null
}
```

#### Invalid Range

- HTTP `400`

```json
{
  "message": "range invalid",
  "result": null
}
```

#### Internal Error

- HTTP `503`

```json
{
  "message": "internal error",
  "result": null
}
```

---

## GET /kaspad/addresses/{address}/balance

Returns the balance entry for a single Kaspa address from the Kaspa gRPC backend.

### Path Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `address` | string | Yes | Kaspa address |

### Success Response

```json
{
  "message": "successful",
  "result": {
    "address": "kaspa:...",
    "balance": 123456789,
    "error": null
  }
}
```

### Result Field Definitions

| Field | Type | Meaning |
|------|------|---------|
| `address` | string | Queried Kaspa address |
| `balance` | uint64 | Current balance for the address |
| `error` | object/null | RPC-level error for this specific address entry, if any |

### Empty Result Response

If the backend returns no entry or more than one entry:

```json
{
  "message": "successful",
  "result": null
}
```

### Error Responses

#### Invalid Address

- HTTP `400`

```json
{
  "message": "address invalid",
  "result": null
}
```

#### Kaspad Error

- HTTP `503`

```json
{
  "message": "kaspad error",
  "result": null
}
```

---

## GET /kaspad/addresses/{address}/utxos

Returns UTXOs for a single Kaspa address from the Kaspa gRPC backend.

### Path Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `address` | string | Yes | Kaspa address |

### Success Response

```json
{
  "message": "successful",
  "result": [
    {
      "address": "kaspa:...",
      "outpoint": {
        "transactionId": "txid",
        "index": 0
      },
      "utxoEntry": {
        "amount": 1000,
        "scriptPublicKey": {
          "version": 0,
          "scriptPublicKey": "..."
        },
        "blockDaaScore": 123456,
        "isCoinbase": false,
        "verboseData": {
          "scriptPublicKeyType": "...",
          "scriptPublicKeyAddress": "kaspa:..."
        }
      }
    }
  ]
}
```

### Result Item Field Definitions

| Field | Type | Meaning |
|------|------|---------|
| `address` | string | Address owning the UTXO |
| `outpoint` | object | Unique reference to the unspent output |
| `utxoEntry` | object | UTXO data itself |

### `outpoint` Field Definitions

| Field | Type | Meaning |
|------|------|---------|
| `transactionId` | string | Transaction ID containing the output |
| `index` | uint32 | Output index inside that transaction |

### `utxoEntry` Field Definitions

| Field | Type | Meaning |
|------|------|---------|
| `amount` | uint64 | Amount stored in the UTXO |
| `scriptPublicKey` | object | Locking script information |
| `blockDaaScore` | uint64 | DAA score of the block associated with the UTXO |
| `isCoinbase` | bool | Whether the UTXO comes from a coinbase transaction |
| `verboseData` | object | Additional decoded script/address information |

### `scriptPublicKey` Field Definitions

| Field | Type | Meaning |
|------|------|---------|
| `version` | uint32 | Script public key version |
| `scriptPublicKey` | string | Script public key content |

### `verboseData` Field Definitions

| Field | Type | Meaning |
|------|------|---------|
| `scriptPublicKeyType` | string | Script type |
| `scriptPublicKeyAddress` | string | Decoded address for the script |

### Empty Result Response

```json
{
  "message": "successful",
  "result": null
}
```

### Error Responses

#### Invalid Address

- HTTP `400`

```json
{
  "message": "address invalid",
  "result": null
}
```

#### Kaspad Error

- HTTP `503`

```json
{
  "message": "kaspad error",
  "result": null
}
```

---

## POST /kaspad/transactions

Submits a new transaction to the Kaspa backend.

### Request Body

JSON body mapped to:

- `protowire.RpcTransaction`

Main request fields visible from the proto:

| Field | Type | Meaning |
|------|------|---------|
| `version` | uint32 | Transaction version |
| `inputs` | array | Transaction inputs |
| `outputs` | array | Transaction outputs |
| `lockTime` | uint64 | Transaction lock time |
| `subnetworkId` | string | Subnetwork ID |
| `gas` | uint64 | Gas value |
| `payload` | string | Transaction payload |
| `verboseData` | object | Optional verbose transaction data |
| `mass` | uint64 | Transaction mass |

### Success Response

If submission succeeds without backend transaction-level error:

```json
{
  "message": "successful",
  "result": {
    "transactionId": "txid",
    "error": null
  }
}
```

### Result Field Definitions

The exact generated response structure depends on `SubmitTransactionResponseMessage`.  
At minimum, the code explicitly checks:

| Field | Type | Meaning |
|------|------|---------|
| `error` | object/null | RPC/backend error for the submitted transaction |

Common implementations also include a transaction identifier such as `transactionId`.

### Failed Submission Response

If the RPC call succeeds but the response contains an error:

```json
{
  "message": "failed",
  "result": {
    "error": {
      "...": "backend error details"
    }
  }
}
```

### Error Responses

#### Invalid Request Body

- HTTP `400`

```json
{
  "message": "data invalid",
  "result": null
}
```

#### Kaspad Error

- HTTP `503`

```json
{
  "message": "kaspad error",
  "result": null
}
```

---

## POST /kaspad/transactions/rbf

Submits a replacement transaction using RBF semantics.

### Request Body

JSON body mapped to:

- `protowire.RpcTransaction`

Main request fields visible from the proto:

| Field | Type | Meaning |
|------|------|---------|
| `version` | uint32 | Transaction version |
| `inputs` | array | Transaction inputs |
| `outputs` | array | Transaction outputs |
| `lockTime` | uint64 | Transaction lock time |
| `subnetworkId` | string | Subnetwork ID |
| `gas` | uint64 | Gas value |
| `payload` | string | Transaction payload |
| `verboseData` | object | Optional verbose transaction data |
| `mass` | uint64 | Transaction mass |

### Success Response

If submission succeeds without backend transaction-level error:

```json
{
  "message": "successful",
  "result": {
    "transactionId": "replacement-txid",
    "error": null
  }
}
```

### Result Field Definitions

The exact generated response structure depends on `SubmitTransactionReplacementResponseMessage`.  
At minimum, the code explicitly checks:

| Field | Type | Meaning |
|------|------|---------|
| `error` | object/null | RPC/backend error for the replacement transaction |

### Failed Submission Response

If the RPC call succeeds but the response contains an error:

```json
{
  "message": "failed",
  "result": {
    "error": {
      "...": "backend error details"
    }
  }
}
```

### Error Responses

#### Invalid Request Body

- HTTP `400`

```json
{
  "message": "data invalid",
  "result": null
}
```

#### Kaspad Error

- HTTP `503`

```json
{
  "message": "kaspad error",
  "result": null
}
```

---

## HTTP Status Summary

| Status Code | Meaning |
|------------|---------|
| `200` | Request succeeded |
| `400` | Invalid path/query/body input |
| `404` | Route not found |
| `500` | Status middleware failed to obtain backend status |
| `503` | Backend unavailable, unsynced, or internal processing error |

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

## Source Notes

This document was regenerated from the current route registration and handlers in:

- `api/init.go`
- `api/routeBookStatus.go`
- `api/routeBookAddress.go`
- `api/routeBookVspc.go`
- `api/routeKaspadAddress.go`
- `api/routeKaspadTransaction.go`
- `api/format.go`
- `database/runtime.go`
- `proto/protowire/rpc.proto`

Some RPC response structures are defined in protobuf and may contain additional fields not directly used by the handlers. This document focuses on fields that can be confirmed from the current repository code.