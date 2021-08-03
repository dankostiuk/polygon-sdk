package jsonrpc

import (
	"fmt"

	"github.com/0xPolygon/minimal/types"
)

// Txpool is the txpool jsonrpc endpoint
type Txpool struct {
	d *Dispatcher
}

type ContentResponse struct {
	Pending   map[types.Address]map[uint64]*gethPendingTransaction    `json:"pending"`
	Queued    map[types.Address]map[uint64]*gethQueuedTransaction     `json:"queued"`
}

type InspectResponse struct {
	Pending   map[string]map[string]string  `json:"pending"`
	Queued		map[string]map[string]string  `json:"queued"`
}

type StatusResponse struct {
	Pending   uint64	`json:"pending"`
	Queued 		uint64	`json:"queued"`
}

type gethPendingTransaction struct {
	Nonce       argUint64      `json:"nonce"`
	GasPrice    argBig         `json:"gasPrice"`
	Gas         argUint64      `json:"gas"`
	To          *types.Address `json:"to"`
	Value       argBig         `json:"value"`
	Input       argBytes       `json:"input"`
	V           argByte        `json:"v"`
	R           argBytes       `json:"r"`
	S           argBytes       `json:"s"`
	Hash        types.Hash     `json:"hash"`
	From        types.Address  `json:"from"`
	BlockHash   interface{}    `json:"blockHash"`
	BlockNumber interface{}    `json:"blockNumber"`
	TxIndex     interface{}    `json:"transactionIndex"`
}

type gethQueuedTransaction struct {
	Nonce       argUint64      `json:"nonce"`
	GasPrice    argBig         `json:"gasPrice"`
	Gas         argUint64      `json:"gas"`
	To          *types.Address `json:"to"`
	Value       argBig         `json:"value"`
	Input       argBytes       `json:"input"`
	V           argByte        `json:"v"`
	R           argBytes       `json:"r"`
	S           argBytes       `json:"s"`
	Hash        types.Hash     `json:"hash"`
	From        types.Address  `json:"from"`
	BlockHash   types.Hash     `json:"blockHash"`
	BlockNumber argUint64      `json:"blockNumber"`
	TxIndex     argUint64      `json:"transactionIndex"`
}

func toGethPendingTransaction(t *types.Transaction, b *types.Block) *gethPendingTransaction {
	if t.R == nil {
		t.R = []byte{0}
	}
	if t.S == nil {
		t.S = []byte{0}
	}
	return &gethPendingTransaction{
		Nonce:       argUint64(t.Nonce),
		GasPrice:    argBig(*t.GasPrice),
		Gas:         argUint64(t.Gas),
		To:          t.To,
		Value:       argBig(*t.Value),
		Input:       argBytes(t.Input),
		V:           argByte(t.V), 
		R:           argBytes(t.R),
		S:           argBytes(t.S),
		Hash:        t.Hash,
		From:        t.From,
		BlockHash: 	 b.Hash(),
		BlockNumber: nil,
		TxIndex: 		 nil,
	}
}

func toGethQueuedTransaction(t *types.Transaction, b *types.Block, txIndex int) *gethQueuedTransaction {
	if t.R == nil {
		t.R = []byte{0}
	}
	if t.S == nil {
		t.S = []byte{0}
	}
	return &gethQueuedTransaction{
		Nonce:       argUint64(t.Nonce),
		GasPrice:    argBig(*t.GasPrice),
		Gas:         argUint64(t.Gas),
		To:          t.To,
		Value:       argBig(*t.Value),
		Input:       argBytes(t.Input),
		V:           argByte(t.V),
		R:           argBytes(t.R), // needs to be 0x0
		S:           argBytes(t.S), // needs to be 0x0
		Hash:        t.Hash,
		From:        t.From,
		BlockHash:   b.Hash(),
		BlockNumber: argUint64(b.Number()),
		TxIndex:     argUint64(txIndex),
	}
}

/** 
 * Content - implemented according to https://geth.ethereum.org/docs/rpc/ns-txpool#txpool_content
 */
func (t *Txpool) Content() (interface{}, error) {
	pendingTxs, queuedTxs := t.d.store.GetTxs()
	pendingRpcTxns := make(map[types.Address]map[uint64]*gethPendingTransaction)
	for address, nonces := range pendingTxs {
		pendingRpcTxns[address] = make(map[uint64]*gethPendingTransaction)
		for nonce, tx := range nonces {
			blockHash, _ := t.d.store.ReadTxLookup(tx.Hash)
			block, _ := t.d.store.GetBlockByHash(blockHash, false)
			// handle genesis block case
			if block == nil {
				block = &types.Block{
					Header: &types.Header{
						Hash: blockHash,
						Number: uint64(0),
					},
				}
			}
			pendingRpcTxns[address][nonce] = toGethPendingTransaction(tx, block)
		}
  }
	queuedRpcTxns := make(map[types.Address]map[uint64]*gethQueuedTransaction)
	for address, nonces := range queuedTxs {
		queuedRpcTxns[address] = make(map[uint64]*gethQueuedTransaction)
		for nonce, tx := range nonces {
			blockHash, _ := t.d.store.ReadTxLookup(tx.Hash)
			block, _ := t.d.store.GetBlockByHash(blockHash, false)
			// handle genesis block case
			if block == nil {
				block = &types.Block{
					Header: &types.Header{
						Hash: blockHash,
						Number: uint64(0),
					},
				}
			}
			// using 0 as txIndex
			queuedRpcTxns[address][nonce] = toGethQueuedTransaction(tx, block, 0)
		}
  }

	resp := ContentResponse{
		Pending: pendingRpcTxns,
		Queued: queuedRpcTxns,
	}
	
	return resp, nil
}

/** 
 * Inspect - implemented according to https://geth.ethereum.org/docs/rpc/ns-txpool#txpool_inspect
 */
func (t *Txpool) Inspect() (interface{}, error) {

	pendingTxs, queuedTxs := t.d.store.GetTxs()
	pendingRpcTxns := make(map[string]map[string]string)
	for address, nonces := range pendingTxs {
		pendingRpcTxns[address.String()] = make(map[string]string)
		for nonce, tx := range nonces {
			msg := fmt.Sprintf("%d wei + %d gas x %d wei", tx.Value, tx.Gas, tx.GasPrice)
			pendingRpcTxns[address.String()][fmt.Sprint(nonce)] = msg
		}
  }

	queuedRpcTxns := make(map[string]map[string]string)
	for address, nonces := range queuedTxs {
		queuedRpcTxns[address.String()] = make(map[string]string)
		for nonce, tx := range nonces {
			msg := fmt.Sprintf("%d wei + %d gas x %d wei", tx.Value, tx.Gas, tx.GasPrice)
			queuedRpcTxns[address.String()][fmt.Sprint(nonce)] = msg
		}
  }

	resp := InspectResponse{
		Pending: pendingRpcTxns,
		Queued: queuedRpcTxns,
	}

	return resp, nil
}

/** 
 * Status - implemented according to https://geth.ethereum.org/docs/rpc/ns-txpool#txpool_content
 */
func (t *Txpool) Status() (interface{}, error) {
	pendingTxs, queuedTxs := t.d.store.GetTxs()
	var pendingCount int
  for _, t := range pendingTxs {
    pendingCount += len(t)
  }
	var queuedCount int
  for _, t := range queuedTxs {
    queuedCount += len(t)
  }

	resp := StatusResponse{
		Pending: uint64(pendingCount),
		Queued: uint64(queuedCount),
	}

	return resp, nil
}