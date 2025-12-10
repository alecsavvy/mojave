package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	accountv1 "github.com/alecsavvy/mojave/gen/account/v1"
	v1 "github.com/alecsavvy/mojave/gen/api/v1"
	chainv1 "github.com/alecsavvy/mojave/gen/chain/v1"
	"github.com/alecsavvy/mojave/sdk"
	"github.com/cosmos/gogoproto/proto"
)

func getNodeURL() string {
	if url := os.Getenv("SONATA_NODE_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
}

// buildCreateAccountTx constructs a signed transaction for creating an account.
func buildCreateAccountTx(account *accountv1.Account) ([]byte, error) {
	signedTx := &chainv1.SignedTransaction{
		Transaction: &chainv1.Transaction{
			Header: &chainv1.TransactionHeader{
				ChainId:   "mojave-test",
				Nonce:     1,
				GasPrice:  1,
				GasLimit:  100000,
				Timeout:   uint64(time.Now().Add(time.Hour).Unix()),
				Sender:    account.Address,
				Recipient: "",
			},
			Body: &chainv1.TransactionBody{
				Body: &chainv1.TransactionBody_CreateAccount{
					CreateAccount: &chainv1.CreateAccountTransaction{
						Account: account,
					},
				},
			},
		},
		Signature: &chainv1.TransactionSignature{
			Signature: []byte("test-signature"), // Placeholder for tests
		},
	}

	return proto.Marshal(signedTx)
}

func TestCreateAndGetAccount(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize SDK
	nodeURL := getNodeURL()
	client := sdk.NewSonataSDK(nodeURL)

	// Create a unique test account
	testAddress := fmt.Sprintf("sonata1test%d", time.Now().UnixNano())
	testAccount := &accountv1.Account{
		Address: testAddress,
		PubKey:  "test-pub-key",
		Balance: 1000,
		Nonce:   0,
	}

	// Build the transaction
	txBytes, err := buildCreateAccountTx(testAccount)
	if err != nil {
		t.Fatalf("failed to build create account transaction: %v", err)
	}

	// Send the transaction
	sendReq := connect.NewRequest(&v1.SendTransactionRequest{
		SignedTransaction: txBytes,
	})

	sendResp, err := client.Chain.SendTransaction(ctx, sendReq)
	if err != nil {
		t.Fatalf("failed to send transaction: %v", err)
	}

	t.Logf("transaction sent, hash: %s", sendResp.Msg.TxHash)

	// Query the account
	getReq := connect.NewRequest(&v1.GetAccountRequest{
		Address: testAddress,
	})

	getResp, err := client.Account.GetAccount(ctx, getReq)
	if err != nil {
		t.Fatalf("failed to get account: %v", err)
	}

	// Verify the account data
	retrievedAccount := getResp.Msg.Account
	if retrievedAccount == nil {
		t.Fatal("retrieved account is nil")
	}
	t.Logf("retrieved account: %s", retrievedAccount.String())

	if retrievedAccount.Address != testAccount.Address {
		t.Errorf("address mismatch: got %s, want %s", retrievedAccount.Address, testAccount.Address)
	}

	if retrievedAccount.PubKey != testAccount.PubKey {
		t.Errorf("pub_key mismatch: got %s, want %s", retrievedAccount.PubKey, testAccount.PubKey)
	}

	if retrievedAccount.Balance != testAccount.Balance {
		t.Errorf("balance mismatch: got %d, want %d", retrievedAccount.Balance, testAccount.Balance)
	}

	if retrievedAccount.Nonce != testAccount.Nonce {
		t.Errorf("nonce mismatch: got %d, want %d", retrievedAccount.Nonce, testAccount.Nonce)
	}

	t.Logf("successfully created and retrieved account: %s", testAddress)
}
