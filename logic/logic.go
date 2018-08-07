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

package logic

import (
	"errors"

	"github.com/yeeshi/go-dappley/client"
	"github.com/yeeshi/go-dappley/core"
	"github.com/yeeshi/go-dappley/storage"
	"github.com/yeeshi/go-dappley/util"
)

var (
	ErrInvalidAddress       = errors.New("ERROR: Address is invalid")
	ErrInvalidSenderAddress = errors.New("ERROR: Sender address is invalid")
	ErrInvalidRcverAddress  = errors.New("ERROR: Receiver address is invalid")
)

//create a blockchain
func CreateBlockchain(address core.Address, db storage.Storage) (*core.Blockchain, error) {
	if !address.ValidateAddress() {
		return nil, ErrInvalidAddress
	}

	bc:= core.CreateBlockchain(address, db)

	return bc, nil
}

//create a wallet
func CreateWallet() (client.Wallet, error) {
	wallets, err := client.NewWallets()
	wallet := wallets.CreateWallet()
	wallets.SaveWalletToFile()

	return wallet, err
}

//get balance
func GetBalance(address core.Address, db storage.Storage) (int, error) {
	if !address.ValidateAddress() {
		return 0, ErrInvalidAddress
	}
	//inject db here

	bc, err := core.GetBlockchain(db)
	if err != nil {
		return 0, err
	}

	balance := 0
	pubKeyHash := util.Base58Decode([]byte(address.Address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs, err := bc.FindUTXO(pubKeyHash)
	if err != nil {
		return 0, err
	}

	for _, out := range UTXOs {
		balance += out.Value
	}
	return balance, nil
}

//get all addresses
func GetAllAddresses() ([]core.Address, error) {
	wallets, err := client.NewWallets()
	if err != nil {
		return nil, err
	}

	addresses := wallets.GetAddresses()

	return addresses, err
}

func Send(from, to core.Address, amount int, tip uint64, db storage.Storage) error {
	if !from.ValidateAddress() {
		return ErrInvalidSenderAddress
	}
	if !to.ValidateAddress() {
		return ErrInvalidRcverAddress
	}

	bc, err := core.GetBlockchain(db)
	if err != nil {
		return err
	}

	wallets, err := client.NewWallets()
	if err != nil {
		return err
	}
	wallet := wallets.GetKeyPairByAddress(from)
	tx, err := core.NewUTXOTransaction(db, from, to, amount, wallet, bc, tip)

	if err != nil {
		return err
	}
	core.GetTxnPoolInstance().Push(tx)

	return err
}

//add balance
func AddBalance(address core.Address, amount int, db storage.Storage) (error) {
	if !address.ValidateAddress() {
		return ErrInvalidAddress
	}
	//inject db here

	bc, err := core.GetBlockchain(db)
	if err != nil {
		return err
	}
	wallets, err := client.NewWallets()
	if err != nil {
		return err
	}
	wallet := wallets.GetKeyPairByAddress(address)
	tx, err := core.NewUTXOTransactionforAddBalance(address, amount, wallet, bc, 0)

	if err != nil {
		return err
	}

	core.GetTxnPoolInstance().Push(tx)

	return err

}


//delete wallet

func DeleteWallet(key *core.KeyPair) error {
	wallets, err := client.NewWallets()
	if err != nil {
		return err
	}
	err = wallets.DeleteWallet(key)
	if err != nil {
		return err
	}
	wallets.SaveWalletToFile()
	return err
}

func DeleteWallets() error {
	wallets, err := client.NewWallets()
	if err != nil {
		return err
	}
	err = wallets.DeleteWallets()
	if err != nil {
		return err
	}
	wallets.SaveWalletToFile()
	return err
}
