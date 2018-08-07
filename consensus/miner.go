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
	"github.com/yeeshi/go-dappley/core"
)

type Miner struct {
	consensus    		core.Consensus
}

//create a new instance
func NewMiner(consensus core.Consensus) *Miner {

	return &Miner{
		consensus:     		consensus,
	}
}

func (miner *Miner) Setup(bc *core.Blockchain, cbAddr string){
	miner.consensus.Setup(bc,cbAddr)
}


//start mining
func (miner *Miner) Start() {
	miner.consensus.Start()
}

func (miner *Miner) Stop() {
	miner.consensus.Stop()
}




