package core

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"log"

	"fmt"

	"github.com/yeeshi/go-dappley/storage"
	logger "github.com/sirupsen/logrus"
)

var tipKey = []byte("1")

const BlockPoolMaxSize = 100

var(
	ErrBlockDoesNotExist			= errors.New("ERROR: Block does not exist in blockchain")
	ErrNotAbleToGetLastBlockHash 	= errors.New("ERROR: Not able to get last block hash in blockchain")
	ErrTransactionNotFound			= errors.New("ERROR: Transaction not found")
)

type Blockchain struct {
	currentHash []byte
	DB          storage.Storage
	blockPool   *BlockPool
	//txPool      *TransactionPool
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address Address, db storage.Storage) *Blockchain {
	genesis := NewGenesisBlock(address.Address)
	bc := &Blockchain{genesis.GetHash(), db,NewBlockPool(BlockPoolMaxSize)}
	bc.blockPool.SetBlockchain(bc)
	bc.updateDbWithNewBlock(genesis)
	return bc
}

func GetBlockchain(db storage.Storage) (*Blockchain, error) {
	var tip []byte
	tip, err := db.Get(tipKey)
	if err != nil {
		return nil, err
	}
	bc := &Blockchain{tip, db,NewBlockPool(BlockPoolMaxSize)}
	bc.blockPool.SetBlockchain(bc)
	return bc, nil
}

func (bc *Blockchain) UpdateNewBlock(newBlock *Block) {
	bc.updateDbWithNewBlock(newBlock)

}

func (bc *Blockchain) BlockPool() *BlockPool {
	return bc.blockPool
}

//TODO: optimize performance
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block, err := bci.Next()
		if err != nil {
			return Transaction{}, err
		}

		for _, tx := range block.GetTransactions() {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.GetPrevHash()) == 0 {
			break
		}
	}

	return Transaction{}, ErrTransactionNotFound
}

//TODO: optimize performance
func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) ([]Transaction, error) {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block, err := bci.Next()
		if err != nil {
		}

		for _, tx := range block.GetTransactions() {
			txID := hex.EncodeToString(tx.ID)

		Outputs: //TODO
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.GetPrevHash()) == 0 {
			break
		}
	}

	return unspentTXs, nil
}

func (bc *Blockchain) FindUTXO(pubKeyHash []byte) ([]TXOutput, error) {
	var UTXOs []TXOutput
	unspentTransactions, err := bc.FindUnspentTransactions(pubKeyHash)
	if err != nil {
		return nil, err
	}

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs, nil
}

func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

func (bc *Blockchain) VerifyTransaction(tx Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err == ErrTransactionNotFound {
			return false
		}
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func (bc *Blockchain) Iterator() *Blockchain {
	return initializeBlockChainWithBlockPool(bc.currentHash, bc.DB)
}

func (bc *Blockchain) Next() (*Block, error) {
	var block *Block

	encodedBlock, err := bc.DB.Get(bc.currentHash)
	if err != nil {
		return nil, err
	}

	block = Deserialize(encodedBlock)

	bc.currentHash = block.GetPrevHash()

	return block, nil
}

func (bc *Blockchain) GetTailHash() ([]byte, error) {

	data, err:= bc.DB.Get(tipKey)
	if err!=nil{
		logger.Error(err)
	}
	return data, err
}

func (bc *Blockchain) String() string {
	var buffer bytes.Buffer

	bci := bc.Iterator()
	for {
		block, err := bci.Next()
		if err != nil {
			fmt.Println(err)
		}

		buffer.WriteString(fmt.Sprintf("============ Block %x ============\n", block.GetHash()))
		buffer.WriteString(fmt.Sprintf("Prev. block: %x\n", block.GetPrevHash()))
		for _, tx := range block.GetTransactions() {
			buffer.WriteString(tx.String())
		}
		buffer.WriteString(fmt.Sprintf("\n\n"))

		if len(block.GetPrevHash()) == 0 {
			break
		}
	}
	return buffer.String()
}

func initializeBlockChainWithBlockPool(current []byte, db storage.Storage) *Blockchain {
	bc := &Blockchain{current, db,NewBlockPool(BlockPoolMaxSize)}
	bc.blockPool.SetBlockchain(bc)
	return bc
}

//record the new block in the database
func (bc *Blockchain) updateDbWithNewBlock(newBlock *Block) {
	bc.DB.Put(newBlock.GetHash(), newBlock.Serialize())
	UpdateUtxoIndexAfterNewBlock(*newBlock, bc.DB)
	bc.setTailBlockHash(newBlock.GetHash())
}

func (bc *Blockchain) setTailBlockHash(hash Hash){
	bc.DB.Put(tipKey, hash)
	bc.currentHash = hash
}

func (bc *Blockchain) GetTailBlock() (*Block, error){
	hash, err:= bc.GetTailHash()
	if err != nil {
		return nil, ErrNotAbleToGetLastBlockHash
	}
	return bc.GetBlockByHash(hash)
}

func (bc *Blockchain) GetMaxHeight() uint64{
	blk, err:= bc.GetTailBlock()
	if err != nil{
		return 0
	}
	return blk.GetHeight()
}

func (bc *Blockchain) HigherThanBlockchain(blk *Block) bool{
	return blk.GetHeight() > bc.GetMaxHeight()
}

func (bc *Blockchain) IsInBlockchain(hash Hash) (bool){
	_, err := bc.GetBlockByHash(hash)
	return err==nil
}

func (bc *Blockchain) MergeFork(){

	//find parent block
	forkHeadBlock := bc.BlockPool().GetForkPoolHeadBlk()
	if forkHeadBlock == nil {
		return
	}
	forkParentHash := forkHeadBlock.GetPrevHash()
	if !bc.IsInBlockchain(forkParentHash){
		return
	}

	//rollback all child blocks after the parent block from tail to head
	bc.RollbackToABlock(forkParentHash)

	//add all blocks in fork from head to tail
	bc.ConcatenateForkToBlockchain()

	logger.Debug("Merge Fork!!")
}

func (bc *Blockchain) ConcatenateForkToBlockchain(){
	if bc.BlockPool().ForkPoolLen() > 0 {
		for i := bc.BlockPool().ForkPoolLen()-1; i>=0; i--{
			bc.UpdateNewBlock(bc.BlockPool().forkPool[i])
			//TODO: Remove transactions in current transaction pool
		}
	}
	bc.BlockPool().ResetForkPool()
}

//returns true if successful
func (bc *Blockchain) RollbackToABlock(hash Hash) bool{

	if !bc.IsInBlockchain(hash){
		return false
	}

	parentBlkHash, err:= bc.GetTailHash()
	if err!= nil {
		return false
	}

	//keep rolling back blocks until the block with the input hash
	loop:
	for {
		if bytes.Compare(parentBlkHash, hash)==0 {
			break loop
		}
		blk,err := bc.GetBlockByHash(parentBlkHash)
		if err!=nil {
			return false
		}
		parentBlkHash = blk.GetPrevHash()
		blk.Rollback()
	}

	bc.setTailBlockHash(parentBlkHash)

	return true
}

func (bc *Blockchain) GetBlockByHash(hash Hash) (*Block, error){
	v, err:=bc.DB.Get(hash)
	if err != nil {
		return nil, ErrBlockDoesNotExist
	}
	return Deserialize(v),nil
}





