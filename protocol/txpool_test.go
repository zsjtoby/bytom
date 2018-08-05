package protocol

import (
	"testing"
	"time"

	"github.com/bytom/consensus"
	"github.com/bytom/protocol/bc"
	"github.com/bytom/protocol/bc/types"
	"github.com/bytom/protocol/vm/vmutil"
	"github.com/bytom/testutil"
)

var testTxs = []*types.Tx{
	types.NewTx(types.TxData{
		SerializedSize: 100,
		TimeRange:      0,
		Inputs: []*types.TxInput{
			types.NewSpendInput(nil, bc.NewHash([32]byte{0x01}), *consensus.BTMAssetID, 1, 1, []byte{0x51}),
		},
		Outputs: []*types.TxOutput{
			types.NewTxOutput(*consensus.BTMAssetID, 1, []byte{0x6a}),
		},
	}),
	types.NewTx(types.TxData{
		SerializedSize: 100,
		TimeRange:      0,
		Inputs: []*types.TxInput{
			types.NewSpendInput(nil, bc.NewHash([32]byte{0x01}), *consensus.BTMAssetID, 1, 1, []byte{0x51}),
		},
		Outputs: []*types.TxOutput{
			types.NewTxOutput(*consensus.BTMAssetID, 1, []byte{0x6b}),
		},
	}),
}

/*func TestTxPool(t *testing.T) {
	p := NewTxPool(nil)

	txA := mockCoinbaseTx(1000, 6543)
	txB := mockCoinbaseTx(2000, 2324)
	txC := mockCoinbaseTx(3000, 9322)

	p.addTransaction(txA, false, 1000, 5000000000)
	if !p.IsTransactionInPool(&txA.ID) {
		t.Errorf("fail to find added txA in tx pool")
	} else {
		i, _ := p.GetTransaction(&txA.ID)
		if i.Height != 1000 || i.Fee != 5000000000 || i.FeePerKB != 5000000000 {
			t.Errorf("incorrect data of TxDesc structure")
		}
	}

	if p.IsTransactionInPool(&txB.ID) {
		t.Errorf("shouldn't find txB in tx pool")
	}
	p.addTransaction(txB, false, 1, 5000000000)
	if !p.IsTransactionInPool(&txB.ID) {
		t.Errorf("shouldn find txB in tx pool")
	}

	if p.Count() != 2 {
		t.Errorf("get wrong number of tx in the pool")
	}
	p.RemoveTransaction(&txB.ID)
	if p.IsTransactionInPool(&txB.ID) {
		t.Errorf("shouldn't find txB in tx pool")
	}

	p.AddErrCache(&txC.ID, nil)
	if !p.IsTransactionInErrCache(&txC.ID) {
		t.Errorf("shouldn find txC in tx err cache")
	}
	if !p.HaveTransaction(&txC.ID) {
		t.Errorf("shouldn find txC in tx err cache")
	}
}*/

func TestRemoveOrphan(t *testing.T) {
	cases := []struct {
		before       *TxPool
		after        *TxPool
		removeHashes []*bc.Hash
	}{
		{
			before: &TxPool{
				orphans: map[bc.Hash]*orphanTx{
					testTxs[0].ID: &orphanTx{
						expiration: time.Unix(1533489701, 0),
						TxDesc: &TxDesc{
							Tx: testTxs[0],
						},
					},
				},
				orphansByPrev: map[bc.Hash]map[bc.Hash]*orphanTx{
					testTxs[0].SpentOutputIDs[0]: map[bc.Hash]*orphanTx{
						testTxs[0].ID: &orphanTx{
							expiration: time.Unix(1533489701, 0),
							TxDesc: &TxDesc{
								Tx: testTxs[0],
							},
						},
					},
				},
			},
			after: &TxPool{
				orphans:       map[bc.Hash]*orphanTx{},
				orphansByPrev: map[bc.Hash]map[bc.Hash]*orphanTx{},
			},
			removeHashes: []*bc.Hash{
				&testTxs[0].ID,
			},
		},
		{
			before: &TxPool{
				orphans: map[bc.Hash]*orphanTx{
					testTxs[0].ID: &orphanTx{
						expiration: time.Unix(1533489701, 0),
						TxDesc: &TxDesc{
							Tx: testTxs[0],
						},
					},
					testTxs[1].ID: &orphanTx{
						expiration: time.Unix(1533489701, 0),
						TxDesc: &TxDesc{
							Tx: testTxs[1],
						},
					},
				},
				orphansByPrev: map[bc.Hash]map[bc.Hash]*orphanTx{
					testTxs[0].SpentOutputIDs[0]: map[bc.Hash]*orphanTx{
						testTxs[0].ID: &orphanTx{
							expiration: time.Unix(1533489701, 0),
							TxDesc: &TxDesc{
								Tx: testTxs[0],
							},
						},
						testTxs[1].ID: &orphanTx{
							expiration: time.Unix(1533489701, 0),
							TxDesc: &TxDesc{
								Tx: testTxs[1],
							},
						},
					},
				},
			},
			after: &TxPool{
				orphans: map[bc.Hash]*orphanTx{
					testTxs[0].ID: &orphanTx{
						expiration: time.Unix(1533489701, 0),
						TxDesc: &TxDesc{
							Tx: testTxs[0],
						},
					},
				},
				orphansByPrev: map[bc.Hash]map[bc.Hash]*orphanTx{
					testTxs[0].SpentOutputIDs[0]: map[bc.Hash]*orphanTx{
						testTxs[0].ID: &orphanTx{
							expiration: time.Unix(1533489701, 0),
							TxDesc: &TxDesc{
								Tx: testTxs[0],
							},
						},
					},
				},
			},
			removeHashes: []*bc.Hash{
				&testTxs[1].ID,
			},
		},
	}

	for i, c := range cases {
		for _, hash := range c.removeHashes {
			c.before.removeOrphan(hash)
		}
		if !testutil.DeepEqual(c.before, c.after) {
			t.Errorf("case %d: got %v want %v", i, c.before, c.after)
		}
	}
}

func mockCoinbaseTx(serializedSize uint64, amount uint64) *types.Tx {
	cp, _ := vmutil.DefaultCoinbaseProgram()
	oldTx := &types.TxData{
		SerializedSize: serializedSize,
		Inputs: []*types.TxInput{
			types.NewCoinbaseInput(nil),
		},
		Outputs: []*types.TxOutput{
			types.NewTxOutput(*consensus.BTMAssetID, amount, cp),
		},
	}
	return &types.Tx{
		TxData: *oldTx,
		Tx:     types.MapTx(oldTx),
	}
}
