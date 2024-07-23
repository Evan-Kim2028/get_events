package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const rpcURL = "https://chainrpc.testnet.mev-commit.xyz/"
const abiURL = "https://raw.githubusercontent.com/primev/mev-commit/v0.4.3/contracts-abi/abi/PreConfCommitmentStore.abi"

var contractAddress = common.HexToAddress("0xCAC68D97a56b19204Dd3dbDC103CB24D47A825A3")

func fetchABI(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func main() {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	abiJSON, err := fetchABI(abiURL)
	if err != nil {
		log.Fatalf("Failed to fetch ABI: %v", err)
	}

	contractABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatalf("Failed to filter logs: %v", err)
	}

	eventSignature := []byte("CommitmentStored(bytes32,address,address,uint256,uint64,bytes32,uint64,uint64,string,string,bytes32,bytes,bytes,uint64,bytes)")
	eventHash := crypto.Keccak256Hash(eventSignature)

	for _, vLog := range logs {
		if len(vLog.Topics) > 0 && vLog.Topics[0] == eventHash {
			fmt.Println("Log:", vLog)

			// Unpack the log into the event structure
			var event struct {
				CommitmentIndex     common.Hash
				Bidder              common.Address
				Commiter            common.Address
				Bid                 *big.Int
				BlockNumber         uint64
				BidHash             common.Hash
				DecayStartTimeStamp uint64
				DecayEndTimeStamp   uint64
				TxnHash             string
				RevertingTxHashes   string
				CommitmentHash      common.Hash
				BidSignature        []byte
				CommitmentSignature []byte
				DispatchTimestamp   uint64
				SharedSecretKey     []byte
			}

			// Decode non-indexed event data
			err := contractABI.UnpackIntoInterface(&event, "CommitmentStored", vLog.Data)
			if err != nil {
				log.Fatalf("Failed to unpack log data: %v", err)
			}

			// Decode indexed event data
			event.CommitmentIndex = common.HexToHash(vLog.Topics[1].Hex())

			fmt.Printf("CommitmentStored Event:\n")
			fmt.Printf("  CommitmentIndex: %s\n", event.CommitmentIndex.Hex())
			fmt.Printf("  Bidder: %s\n", event.Bidder.Hex())
			fmt.Printf("  Commiter: %s\n", event.Commiter.Hex())
			fmt.Printf("  Bid: %s\n", event.Bid.String())
			fmt.Printf("  BlockNumber: %d\n", event.BlockNumber)
			fmt.Printf("  BidHash: %s\n", event.BidHash.Hex())
			fmt.Printf("  DecayStartTimeStamp: %d\n", event.DecayStartTimeStamp)
			fmt.Printf("  DecayEndTimeStamp: %d\n", event.DecayEndTimeStamp)
			fmt.Printf("  TxnHash: %s\n", event.TxnHash)
			fmt.Printf("  RevertingTxHashes: %s\n", event.RevertingTxHashes)
			fmt.Printf("  CommitmentHash: %s\n", event.CommitmentHash.Hex())
			fmt.Printf("  BidSignature: %x\n", event.BidSignature)
			fmt.Printf("  CommitmentSignature: %x\n", event.CommitmentSignature)
			fmt.Printf("  DispatchTimestamp: %d\n", event.DispatchTimestamp)
			fmt.Printf("  SharedSecretKey: %x\n", event.SharedSecretKey)
		}
	}
}
