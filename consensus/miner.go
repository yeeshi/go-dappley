// Copyright (C) 2018 go-dappley authors
//
// This file is part of the go-dappley library.
//
// the go-dappley library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-dappley library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-dappley library.  If not, see <http://www.gnu.org/licenses/>.
//
package consensus

import (
	"github.com/dappley/go-dappley/core"
	"container/heap"
	"fmt"
	"github.com/dappley/go-dappley/network"
)

type state int

const (
	prepareTxPoolState state = iota
	mineState
	updateNewBlock
	cleanUpState
)

type Miner struct {
	bc           *core.Blockchain
	newBlock     *core.Block
	coinBaseAddr string
	nextState    state
	consensus    core.Consensus

	exitCh           chan bool
	messageCh        chan string
	newBlockReceived bool
	node             *network.Node
}

//create a new instance
func NewMiner(bc *core.Blockchain, coinBaseAddr string, consensus core.Consensus) *Miner {

	return &Miner{
		bc:           bc,
		newBlock:     nil,
		coinBaseAddr: coinBaseAddr,
		nextState:    prepareTxPoolState,
		consensus:    consensus,
		exitCh:       make(chan bool, 1),
		messageCh:    make(chan string, 128),
		node:         network.NewNode(bc),
	}
}

//start mining
func (miner *Miner) Start() {
	miner.stateLoop()
	miner.messageLoop()
}

func (miner *Miner) Stop() {
	miner.exitCh <- true
}

func (miner *Miner) Feed(msg string) {
	miner.messageCh <- msg
}

func (miner *Miner) FeedBlock(blk *core.Block) {
	miner.bc.BlockPool().Push(blk)
}

func (miner *Miner) stateLoop() {

Loop:
	for {
		switch miner.nextState {
		case prepareTxPoolState:
			miner.prepareTxPool()
			miner.nextState = mineState

		case mineState:
			miner.mine()
			miner.nextState = updateNewBlock
		case updateNewBlock:
			miner.updateNewBlock()
			miner.nextState = cleanUpState
		case cleanUpState:
			miner.cleanUp()
			break Loop
		}
	}
}
func (miner *Miner) messageLoop() {
	for {
		fmt.Println("running")
		select {
		case msg := <-miner.messageCh:
			fmt.Println(msg)
		case block := <-miner.bc.BlockPool().BlockReceivedCh():
			miner.newBlockReceived = true
			miner.newBlock = block
			miner.nextState = updateNewBlock
		case <-miner.exitCh:
			return
		}
	}
}

//start mining
func (pd *Miner) StartMining(signal chan bool) {
Loop:
	for {
		select {
		case stop := <-signal:
			if stop {
				break Loop
			}
		default:
			switch pd.nextState {
			case prepareTxPoolState:
				pd.prepareTxPool()
				pd.nextState = mineState
			case mineState:
				pd.mine()
				pd.nextState = updateNewBlock
			case updateNewBlock:
				pd.updateNewBlock()
				pd.nextState = cleanUpState
			case cleanUpState:
				pd.cleanUp()
				pd.nextState = prepareTxPoolState
			}
		}
	}
}

//prepare transaction pool
func (miner *Miner) prepareTxPool() {
	// verify all transactions
	miner.verifyTransactions()
}

//start proof of work process
func (miner *Miner) mine() {

	//get the hash of last newBlock
	lastHash, err := miner.bc.GetLastHash()
	if err != nil {
		//TODO
	}
	//create a new newBlock with the transaction pool and last hash

	miner.consensus = NewProofOfWork(miner.bc)
	miner.newBlock = miner.consensus.ProduceBlock(miner.coinBaseAddr, "", lastHash)
	miner.node.SendBlock(miner.newBlock)
}

//update the blockchain with the new block
func (miner *Miner) updateNewBlock() {
	miner.bc.UpdateNewBlock(miner.newBlock)
}

func (miner *Miner) cleanUp() {
	miner.nextState = prepareTxPoolState
}

//verify transactions and remove invalid transactions
func (miner *Miner) verifyTransactions() {
	txnPool := core.GetTxnPoolInstance()
	txnPoolLength := txnPool.Len()
	for i := 0; i < txnPoolLength; i++ {
		var txn = heap.Pop(txnPool).(core.Transaction)
		if miner.bc.VerifyTransaction(txn) == true {
			//Remove transaction from transaction pool if the transaction is not verified
			txnPool.Push(txn)
		}
	}

}
