package core

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/common"
	// "github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/icstglobal/go-icst/chain"
	"github.com/icstglobal/go-icst/chain/eth"
	"github.com/icstglobal/plasma/core/types"

	"reflect"
	"time"
)

const (
	DepositEventName     = "Deposit"
	ExitedStartEventName = "ExistedStart"
)

type RootChain struct {
	chain    chain.Chain
	sub      map[string]func(event *chain.ContractEvent) // map topic0 to name
	abiStr   string
	cxAddr   string
	operator *Operator
}

type DepositEvent struct {
	Depositor    common.Address
	DepositBlock *big.Int
	Token        common.Address
	Amount       *big.Int
	Raw          ethtypes.Log // Blockchain specific contextual infos
}

// chain
func NewRootChain(url string, abiStr string, cxAddr string, operator *Operator) (*RootChain, error) {

	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect eth rpc endpoint {%v}, err is:%v", url, err)
	}
	blc := eth.NewChainEthereum(client)
	chain.Set(chain.Eth, blc)
	rc := &RootChain{
		chain:    blc,
		sub:      make(map[string]func(event *chain.ContractEvent)),
		abiStr:   abiStr,
		cxAddr:   cxAddr,
		operator: operator,
	}
	// register dealing func
	rc.sub[DepositEventName] = rc.dealWithDepositEvent
	rc.sub[ExitedStartEventName] = rc.dealWithExitStartedEvent

	var depositEvent DepositEvent
	// start loop to sync root chain event
	go rc.loopEvent(DepositEventName, depositEvent)
	// go rc.loopEvent(ExitedStartEventName)

	return rc, nil
}

func (rc *RootChain) loopEvent(eventName string, event interface{}) {
	fromBlock := big.NewInt(100)

	cxAddrBytes, err := hex.DecodeString(rc.cxAddr)
	if err != nil {
		log.Error("Decode cxAddr Error:", err)
		return
	}

	for {
		events, err := rc.chain.GetContractEvents(context.Background(), cxAddrBytes, fromBlock, nil, rc.abiStr, eventName, reflect.TypeOf(event))
		if err != nil {
			log.Errorf("chain.GetEvents: %v", err)
			return
		}

		for _, event := range events {
			rc.sub[eventName](event)
		}
		time.Sleep(time.Second * 2)
		if len(events) > 0 {
			fromBlock = big.NewInt(int64(events[len(events)-1].BlockNum) + 1)
		}
	}
}

func (rc *RootChain) dealWithDepositEvent(event *chain.ContractEvent) {
	out := event.V.(*DepositEvent)

	// construct tx
	txOut1 := &types.TxOut{Owner: out.Depositor, Amount: out.Amount}
	txOut2 := &types.TxOut{Owner: common.Address{}, Amount: big.NewInt(0)}
	txIn1 := &types.UTXO{
		UTXOID: types.UTXOID{BlockNum: 0, TxIndex: 0, OutIndex: 0},
		TxOut:  types.TxOut{Owner: common.Address{}, Amount: big.NewInt(0)},
	}
	txIn2 := &types.UTXO{
		UTXOID: types.UTXOID{BlockNum: 0, TxIndex: 0, OutIndex: 0},
		TxOut:  types.TxOut{Owner: common.Address{}, Amount: big.NewInt(0)},
	}

	fee := big.NewInt(1) // todo:fee
	tx := types.NewTransaction(txIn1, txIn2, txOut1, txOut2, fee)

	txs := make(types.Transactions, 0)
	txs = append(txs, tx)
	rc.operator.AddTxs(txs)

	log.Debugf("dealWithDepositEvent: %v blockNumber: %v", out.Depositor.Hex(), event.BlockNum)
}

func (rc *RootChain) dealWithExitStartedEvent(event *chain.ContractEvent) {
}
