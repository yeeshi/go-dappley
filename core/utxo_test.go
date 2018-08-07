package core

import (
	"testing"
	"github.com/yeeshi/go-dappley/storage"
	"time"
	"github.com/yeeshi/go-dappley/util"
	"github.com/stretchr/testify/assert"

)


func GenerateUtxoMockBlockWithoutInputs() *Block{
	bh1 := &BlockHeader{
		[]byte("hash"),
		[]byte("prevhash"),
		1,
		time.Now().Unix(),
	}

	t1 := MockUtxoTransactionWithoutInputs()
	return &Block{
		header:       bh1,
		transactions: []*Transaction{t1},
		height:       0,
	}
}


func GenerateUtxoMockBlockWithInputs() *Block{
	bh1 := &BlockHeader{
		[]byte("hash"),
		[]byte("prevhash"),
		1,
		time.Now().Unix(),
	}

	t1 := MockUtxoTransactionWithInputs()
	return &Block{
		header:       bh1,
		transactions: []*Transaction{t1},
		height:       0,
	}
}

func MockUtxoTransactionWithoutInputs() *Transaction{
	return &Transaction{
		ID:  []byte("txn1"),
		Vin:  []TXInput{},
		Vout: MockUtxoOutputsWithoutInputs(),
		Tip:  5,
	}
}

func MockUtxoTransactionWithInputs() *Transaction{
	return &Transaction{
		ID:   []byte("txn2"),
		Vin:  MockUtxoInputs(),
		Vout: MockUtxoOutputsWithInputs(),
		Tip:  5,
	}
}

func MockUtxoInputs() []TXInput {
	return []TXInput{
		{[]byte("txn1"),
			0,
			util.GenerateRandomAoB(2),
			[]byte("address1")},
		{[]byte("txn1"),
			1,
			util.GenerateRandomAoB(2),
			[]byte("address1")},
	}
}

func MockUtxoOutputsWithoutInputs() []TXOutput {
	return []TXOutput{
		{5, []byte("address1")},
		{7, []byte("address1")},
	}
}

func MockUtxoOutputsWithInputs() []TXOutput {
	return []TXOutput{
		{4, []byte("address1")},
		{5, []byte("address2")},
		{3, []byte("address2")},
	}
}



func TestAddSpendableOutputsAfterNewBlock(t *testing.T){
	db :=  storage.NewRamStorage()
	defer db.Close()
	blk := GenerateUtxoMockBlockWithoutInputs()

	AddSpendableOutputsAfterNewBlock(*blk, db)
	myUtxos := GetAddressUTXOs([]byte("address1"), db)
	assert.Equal(t, 5, myUtxos[0].Value )
	assert.Equal(t, 7, myUtxos[1].Value )
}

func TestConsumeSpentOutputsAfterNewBlock(t *testing.T){
	db :=  storage.NewRamStorage()
	defer db.Close()

	blk1 := GenerateUtxoMockBlockWithoutInputs()
	AddSpendableOutputsAfterNewBlock(*blk1, db)
	//address 1 is given a $5 utxo and a $7 utxo, total $12

	blk2 := GenerateUtxoMockBlockWithInputs()
	//consume utxos first, not adding new utxos yet
	ConsumeSpendableOutputsAfterNewBlock(*blk2, db)
	//address1 gives address2 $8, $12 - $8 = $4 but address1 has no utxos left at this point new(change) utxo hasnt been added
	assert.Equal(t, 0, len( GetAddressUTXOs([]byte("address1"), db)))

	//add utxos for above block accordingly;
	AddSpendableOutputsAfterNewBlock(*blk2, db)

	//expect address1 to have 1 utxo of $4
	assert.Equal(t, 1, len( GetAddressUTXOs([]byte("address1"), db)))
	assert.Equal(t, 4,  GetAddressUTXOs([]byte("address1"), db)[0].Value)

	//expect address2 to have 2 utxos totaling $8
	assert.Equal(t, 2, len( GetAddressUTXOs([]byte("address2"), db)))
	sum := 0
	for _, utxo := range GetAddressUTXOs([]byte("address2"),db) {
		sum += utxo.Value
	}
	assert.Equal(t, 8, sum)

}