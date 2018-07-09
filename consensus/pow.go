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
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"

	"github.com/dappley/go-dappley/core"
	"github.com/dappley/go-dappley/util"
	"container/heap"
	"github.com/copernet/copernicus/log"
)

var maxNonce int64 = math.MaxInt64

const targetBits = int64(14)

type ProofOfWork struct {

	target 		*big.Int

	exitCh chan bool
	messageCh chan string
}

func NewProofOfWork() *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	return &ProofOfWork{target, make(chan bool, 1), make(chan string, 128)}
}


func prepareData(nonce int64, blk *core.Block) []byte {
	data := bytes.Join(
		[][]byte{
			blk.GetPrevHash(),
			blk.HashTransactions(),
			util.IntToHex(blk.GetTimestamp()),
			util.IntToHex(targetBits),
			util.IntToHex(nonce),
		},
		[]byte{},
	)
	return data
}

func (pow *ProofOfWork) ProduceBlock(cbAddr, cbData string, prevHash []byte) *core.Block{

	var hashInt big.Int
	var hash [32]byte
	nonce := int64(0)

	//add coinbase transaction to transaction pool

	cbtx := core.NewCoinbaseTX(cbAddr,cbData)
	h := core.GetTxnPoolInstance()

	heap.Init(h)
	heap.Push(core.GetTxnPoolInstance(), cbtx)

	//prepare the new block (without the correct nonce value)
	blk := core.NewBlock(prevHash)

	//find the nonce value
	fmt.Println("Mining a new block")
	for nonce < maxNonce {
		data := prepareData(nonce, blk)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		}

		nonce++
	}
	fmt.Print("\n\n")

	//complete the block
	blk.SetHash(hash[:])
	blk.SetNonce(nonce)

	return blk
}

func (pow *ProofOfWork) Validate(blk *core.Block) bool {
	var hashInt big.Int

	data := prepareData(blk.GetNonce(), blk)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}

func (pow *ProofOfWork) Stop() {
	pow.exitCh <- true
}

func (pow *ProofOfWork) Feed(msg string) {
	pow.messageCh <- msg
}

func (pow *ProofOfWork) Start() {
	for {
		log.Info("running")
		select {
		case msg:= <-pow.messageCh:
			log.Info(msg)
		case <-pow.exitCh:
			log.Info("quit Pow.")
			return
		}
	}
}
