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

The handler zeroes out `TotalBlock` and `TotalTransaction` before returning the response.

### Response

- `200 OK` when status is available
- `503 Service Unavailable` when Kaspad is not synced

### Example Response

```json
{
  "message": "synced",
  "result": {
    "StatusKaspad": "synced",
    "GapBook": "0",
    "Hysteresis": "100",
    "TotalBlock": 0,
    "TotalTransaction": 0
  }
}
```

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
    "...": "block fields"
  }
}
```

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
    "...": "transaction fields"
  }
}
```

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
      "...": "vspc fields"
    }
  ]
}
```

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
      "...": "vspc fields"
    }
  ]
}
```

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
      "...": "formatted transaction fields"
    }
  ]
}
```

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
    "...": "RpcBalancesByAddressesEntry fields"
  }
}
```

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
      "...": "RpcUtxosByAddressesEntry fields"
    }
  ]
}
```

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

### Success Response

If submission succeeds without backend transaction-level error:

```json
{
  "message": "successful",
  "result": {
    "...": "SubmitTransactionResponseMessage fields"
  }
}
```

### Failed Submission Response

If the RPC call succeeds but the response contains an error:

```json
{
  "message": "failed",
  "result": {
    "...": "SubmitTransactionResponseMessage fields including error"
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

### Success Response

If submission succeeds without backend transaction-level error:

```json
{
  "message": "successful",
  "result": {
    "...": "SubmitTransactionReplacementResponseMessage fields"
  }
}
```

### Failed Submission Response

If the RPC call succeeds but the response contains an error:

```json
{
  "message": "failed",
  "result": {
    "...": "SubmitTransactionReplacementResponseMessage fields including error"
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
- `api/routeBookAddress.go`
- `api/routeBookVspc.go`
- `api/routeKaspadAddress.go`
- `api/routeKaspadTransaction.go`

Some response field schemas such as `formatBlockType`, `formatTransactionType`, `formatVspcType`, and `protowire.*` structures are referenced by the handlers but are defined elsewhere. If needed, those can be expanded into a full schema-level API reference as a next step.