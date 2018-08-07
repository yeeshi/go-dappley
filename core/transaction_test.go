package core

import (
	"testing"

	"github.com/yeeshi/go-dappley/core/pb"
	"github.com/yeeshi/go-dappley/util"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func getAoB(length int64) []byte {
	return util.GenerateRandomAoB(length)
}

func GenerateFakeTxInputs() []TXInput {
	return []TXInput{
		{getAoB(2), 10, getAoB(2), getAoB(2)},
		{getAoB(2), 5, getAoB(2), getAoB(2)},
	}
}

func GenerateFakeTxOutputs() []TXOutput {
	return []TXOutput{
		{1, getAoB(2)},
		{2, getAoB(2)},
	}
}

func TestTrimmedCopy(t *testing.T) {
	var t1 = Transaction{
		ID:   util.GenerateRandomAoB(1),
		Vin:  GenerateFakeTxInputs(),
		Vout: GenerateFakeTxOutputs(),
		Tip:  2,
	}

	t2 := t1.TrimmedCopy()

	t3 := NewCoinbaseTX("13ZRUc4Ho3oK3Cw56PhE5rmaum9VBeAn5F", "")
	t4 := t3.TrimmedCopy()
	assert.Equal(t, t1.ID, t2.ID)
	assert.Equal(t, t1.Tip, t2.Tip)
	assert.Equal(t, t1.Vout, t2.Vout)
	for index, vin := range t2.Vin {
		assert.Nil(t, vin.Signature)
		assert.Nil(t, vin.PubKey)
		assert.Equal(t, t1.Vin[index].Txid, vin.Txid)
		assert.Equal(t, t1.Vin[index].Vout, vin.Vout)
	}

	assert.Equal(t, t3.ID, t4.ID)
	assert.Equal(t, t3.Tip, t4.Tip)
	assert.Equal(t, t3.Vout, t4.Vout)
	for index, vin := range t4.Vin {
		assert.Nil(t, vin.Signature)
		assert.Nil(t, vin.PubKey)
		assert.Equal(t, t3.Vin[index].Txid, vin.Txid)
		assert.Equal(t, t3.Vin[index].Vout, vin.Vout)
	}
}

//test IsCoinBase and NewCoinbaseTX function
func TestIsCoinBase(t *testing.T) {
	var t1 = Transaction{
		ID:   util.GenerateRandomAoB(1),
		Vin:  GenerateFakeTxInputs(),
		Vout: GenerateFakeTxOutputs(),
		Tip:  2,
	}

	assert.False(t, t1.IsCoinbase())

	t2 := NewCoinbaseTX("13ZRUc4Ho3oK3Cw56PhE5rmaum9VBeAn5F", "")

	assert.True(t, t2.IsCoinbase())

}

func TestTransaction_Proto(t *testing.T) {
	t1 := Transaction{
		ID:   util.GenerateRandomAoB(1),
		Vin:  GenerateFakeTxInputs(),
		Vout: GenerateFakeTxOutputs(),
		Tip:  5,
	}

	pb := t1.ToProto()
	var i interface{} = pb
	_, correct := i.(proto.Message)
	assert.Equal(t, true, correct)
	mpb, err := proto.Marshal(pb)
	assert.Nil(t, err)

	newpb := &corepb.Transaction{}
	err = proto.Unmarshal(mpb, newpb)
	assert.Nil(t, err)

	t2 := Transaction{}
	t2.FromProto(newpb)

	assert.Equal(t, t1, t2)
}
