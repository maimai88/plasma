package types

import (
	"io/ioutil"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestFee(t *testing.T) {
	tx := new(Transaction)
	tx.data.Fee = big.NewInt(10)

	fcopy := tx.Fee()
	fcopy = fcopy.Add(fcopy, big.NewInt(1))

	if fcopy.Cmp(tx.data.Fee) == 0 {
		t.Logf("fee:%v, fcopy:%v", tx.data.Fee, fcopy)
		t.Fail()
	}
}

func BenchmarkBigIntCopyByAdd(b *testing.B) {
	tx := new(Transaction)
	tx.data.Fee, _ = new(big.Int).SetString("100000000222222222222222222220000000000000000000", 10)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tx.Fee()
	}
}

func TestMarshalTx(t *testing.T) {
	senderKey, _ := crypto.GenerateKey()
	sender := crypto.PubkeyToAddress(senderKey.PublicKey)

	receiverKey, _ := crypto.GenerateKey()
	receiver := crypto.PubkeyToAddress(receiverKey.PublicKey)

	signer := NewEIP155Signer(big.NewInt(1))

	tx1 := &Transaction{}
	in1 := &UTXO{UTXOID: UTXOID{BlockNum: 1, TxIndex: 0, OutIndex: 0}, TxOut: TxOut{Owner: sender, Amount: big.NewInt(50)}}
	in2 := &UTXO{UTXOID: UTXOID{BlockNum: 1, TxIndex: 1, OutIndex: 0}, TxOut: TxOut{Owner: sender, Amount: big.NewInt(0)}}
	tx1.data.Ins = [2]*UTXO{in1, in2}

	out1 := &TxOut{Owner: receiver, Amount: big.NewInt(90)}
	out2 := &TxOut{Owner: sender, Amount: big.NewInt(0)} //zero output
	tx1.data.Outs = [2]*TxOut{out1, out2}
	tx1.data.Fee = big.NewInt(10)

	tx1, err := SignTx(tx1, signer, senderKey, senderKey)
	if err != nil {
		t.Fatal("failed to sign tx", err)
	}

	buf, err := tx1.marshalJSON(true)
	t.Log(string(buf))
	if err != nil {
		t.Fatal("unable to marshal tx")
	}

	tx2 := new(Transaction)
	err = tx2.UnmarshalJSON(buf)
	if err != nil {
		t.Logf("json: %v", string(buf))
		t.Fatal("unable to unmarshal tx", err)
	}
	if reflect.DeepEqual(tx1, tx2) {
		t.Logf("json: %v", string(buf))
		t.Fatal("tx not correctly marshaled")
	}
}

func TestMarshalTxWithKeyStore(t *testing.T) {

	// load from keystore
	keystoreDir := "../../keystore"
	files, err := ioutil.ReadDir(keystoreDir)
	if err != nil {
		t.Fatal("keystore ioutil.ReadDir Error:", err)
	}

	var key *keystore.Key
	for _, file := range files {
		path := keystoreDir + "/" + file.Name()
		// log.Debug("key path:", path, config.OperPwd)
		jsonStr, err := ioutil.ReadFile(path)
		pwd := "wzp"
		key, err = keystore.DecryptKey(jsonStr, pwd)

		if err != nil {
			t.Fatal(err)
		}
		break
	}
	if key == nil {
		t.Fatal(err)
	}
	sender := crypto.PubkeyToAddress(key.PrivateKey.PublicKey)

	receiverKey, _ := crypto.GenerateKey()
	receiver := crypto.PubkeyToAddress(receiverKey.PublicKey)

	signer := NewEIP155Signer(big.NewInt(1))

	tx1 := &Transaction{}
	in1 := &UTXO{UTXOID: UTXOID{BlockNum: 1, TxIndex: 0, OutIndex: 0}, TxOut: TxOut{Owner: sender, Amount: big.NewInt(10)}}
	in2 := &UTXO{UTXOID: UTXOID{BlockNum: 2, TxIndex: 0, OutIndex: 0}, TxOut: TxOut{Owner: sender, Amount: big.NewInt(10)}}
	tx1.data.Ins = [2]*UTXO{in1, in2}

	out1 := &TxOut{Owner: receiver, Amount: big.NewInt(14)}
	out2 := &TxOut{Owner: sender, Amount: big.NewInt(5)} //zero output
	tx1.data.Outs = [2]*TxOut{out1, out2}
	tx1.data.Fee = big.NewInt(1)

	tx1, err = SignTx(tx1, signer, key.PrivateKey, key.PrivateKey)
	if err != nil {
		t.Fatal("failed to sign tx", err)
	}

	buf, err := tx1.marshalJSON(true)
	t.Log(string(buf))
	if err != nil {
		t.Fatal("unable to marshal tx")
	}

	tx2 := new(Transaction)
	err = tx2.UnmarshalJSON(buf)
	if err != nil {
		t.Logf("json: %v", string(buf))
		t.Fatal("unable to unmarshal tx", err)
	}
	if reflect.DeepEqual(tx1, tx2) {
		t.Logf("json: %v", string(buf))
		t.Fatal("tx not correctly marshaled")
	}
}
