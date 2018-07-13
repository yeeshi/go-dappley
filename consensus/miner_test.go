package consensus

import (
	"os"
	"testing"

	"fmt"
	"time"

	"github.com/dappley/go-dappley/client"
	"github.com/dappley/go-dappley/core"
	"github.com/dappley/go-dappley/storage"
	"github.com/dappley/go-dappley/util"
	"github.com/stretchr/testify/assert"
)

var sendAmount = int(5)
var mineReward = int(10)
var tip = int64(5)

//mine one transaction
func TestMiner_SingleValidTx(t *testing.T) {

	setup()

	//create new wallet

	wallets, err := client.NewWallets()
	assert.Nil(t, err)
	assert.NotNil(t, wallets)

	wallet1 := wallets.CreateWallet()
	assert.NotNil(t, wallet1)

	wallet2 := wallets.CreateWallet()
	assert.NotNil(t, wallet2)

	wallet := wallets.GetKeyPairByAddress(wallet1.GetAddress())

	//create a blockchain
	assert.Equal(t, true, wallet1.GetAddress().ValidateAddress())

	db := storage.OpenDatabase(core.BlockchainDbFile)
	defer db.Close()

	bc, err := core.CreateBlockchain(wallet1.GetAddress(), db)
	assert.Nil(t, err)

	assert.NotNil(t, bc)

	//check balance
	checkBalance(t, wallet1.GetAddress().Address, wallet2.GetAddress().Address, bc, mineReward, 0)

	//create 2 transactions and start mining
	tx, err := core.NewUTXOTransaction(wallet1.GetAddress(), wallet2.GetAddress(), sendAmount, wallet, bc, tip)
	assert.Nil(t, err)

	core.GetTxnPoolInstance().Push(tx)

	miner := NewMiner(bc, wallet1.GetAddress().Address, NewProofOfWork(bc))
	go miner.Start()
	for i := 0; i < 3; i++ {
		miner.Feed(time.Now().String())
		miner.Feed("test test")
		time.Sleep(1 * time.Second)
	}
	miner.Stop()

	checkBalance(t, wallet1.GetAddress().Address, wallet2.GetAddress().Address, bc, mineReward*2-sendAmount, sendAmount)

	teardown()
}

//mine empty blocks
func TestMiner_MineEmptyBlock(t *testing.T) {

	setup()

	//create new wallet
	wallets, _ := client.NewWallets()
	assert.NotNil(t, wallets)

	wallet1 := wallets.CreateWallet()
	assert.NotNil(t, wallet1)

	wallet2 := wallets.CreateWallet()
	assert.NotNil(t, wallet2)

	//create a blockchain
	assert.Equal(t, true, wallet1.GetAddress().ValidateAddress())

	db := storage.OpenDatabase(core.BlockchainDbFile)
	defer db.Close()

	bc, err := core.CreateBlockchain(wallet1.GetAddress(), db)
	assert.Nil(t, err)
	assert.NotNil(t, bc)

	//check balance
	checkBalance(t, wallet1.GetAddress().Address, wallet2.GetAddress().Address, bc, mineReward, 0)

	//create 2 transactions and start mining

	miner := NewMiner(bc, wallet1.GetAddress().Address, NewProofOfWork(bc))
	go miner.Start()
	for i := 0; i < 1; i++ {
		miner.Feed(time.Now().String())
		time.Sleep(1 * time.Second)
	}
	miner.Stop()
	//fmt.Println(bc)
	checkBalance(t, wallet1.GetAddress().Address, wallet2.GetAddress().Address, bc, mineReward*2, 0)

	teardown()
}

//mine multiple transactions
func TestMiner_MultipleValidTx(t *testing.T) {

	setup()

	//create new wallet
	wallets, _ := client.NewWallets()
	assert.NotNil(t, wallets)

	wallet1 := wallets.CreateWallet()
	assert.NotNil(t, wallet1)

	wallet2 := wallets.CreateWallet()
	assert.NotNil(t, wallet2)

	wallet := wallets.GetKeyPairByAddress(wallet1.GetAddress())

	//create a blockchain
	assert.Equal(t, true, wallet1.GetAddress().ValidateAddress())

	db := storage.OpenDatabase(core.BlockchainDbFile)
	defer db.Close()

	bc, err := core.CreateBlockchain(wallet1.GetAddress(), db)
	assert.Nil(t, err)
	assert.NotNil(t, bc)

	//check balance ; a:10, b:0
	checkBalance(t, wallet1.GetAddress().Address, wallet2.GetAddress().Address, bc, mineReward, 0)

	tx, err := core.NewUTXOTransaction(wallet1.GetAddress(), wallet2.GetAddress(), 4, wallet, bc, tip)
	assert.Nil(t, err)

	//a:15 b:5
	core.GetTxnPoolInstance().Push(tx)
	//a:20 b:10

	miner := NewMiner(bc, wallet1.GetAddress().Address, NewProofOfWork(bc))
	go miner.Start()
	for i := 0; i < 1; i++ {
		miner.Feed(time.Now().String())
		time.Sleep(1 * time.Second)
	}

	core.GetTxnPoolInstance().Push(tx)
	for i := 0; i < 1; i++ {
		miner.Feed(time.Now().String())
		time.Sleep(1 * time.Second)
	}
	tx2, err := core.NewUTXOTransaction(wallet1.GetAddress(), wallet2.GetAddress(), 11, wallet, bc, tip)
	core.GetTxnPoolInstance().Push(tx2)
	for i := 0; i < 1; i++ {
		miner.Feed(time.Now().String())
		time.Sleep(1 * time.Second)
	}

	miner.Stop()
	go miner.Start()
	for i := 0; i < 1; i++ {
		miner.Feed(time.Now().String())
		time.Sleep(1 * time.Second)
	}
	miner.Stop()

	fmt.Println(bc)
	checkBalance(t, wallet1.GetAddress().Address, wallet2.GetAddress().Address, bc, 11, 19)

	teardown()

}

//TODO: test mining with invalid transactions
func TestMiner_InvalidTransactions(t *testing.T) {

}

//balance
func getBalance(bc *core.Blockchain, addr string) (int, error) {

	balance := 0
	pubKeyHash := util.Base58Decode([]byte(addr))
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

func setup() {
	cleanUpDatabase()
}

func teardown() {
	cleanUpDatabase()
}

func cleanUpDatabase() {
	os.RemoveAll("../bin/blockchain.DB")
	os.RemoveAll(client.WalletFile)
}

func checkBalance(t *testing.T, addr1, addr2 string, bc *core.Blockchain, addr1v, addr2v int) {
	//check balance after transaction
	balance1, err := getBalance(bc, addr1)
	assert.Nil(t, err)
	assert.Equal(t, addr1v, balance1)

	balance2, err := getBalance(bc, addr2)
	assert.Nil(t, err)
	assert.Equal(t, addr2v, balance2)
}
