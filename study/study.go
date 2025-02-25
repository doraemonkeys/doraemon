package study

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// StrLen统计字符串长度
func StrLen(str string) int {
	return len([]rune(str))
}

func GenRandomETHKey() (privateKey *ecdsa.PrivateKey, err error) {
	return crypto.GenerateKey()
}

// CalculateSelector calculates the Solidity function selector.
func CalculateSelector(signature string) string {
	// 1. Calculate the Keccak-256 hash of the signature.
	hash := crypto.Keccak256Hash([]byte(signature))

	// 2. Take the first 4 bytes (8 hex characters) of the hash.
	selector := hash.Hex()[:10] // Include "0x" prefix

	return selector
}

// ExtractCalldataChunks extracts and formats the calldata (input data) of an Ethereum transaction.
// It retrieves the transaction by its hash, extracts the calldata, and then splits it into 32-byte (64 hexadecimal characters) chunks.
func ExtractCalldataChunks(ctx context.Context, client *ethclient.Client, txHash common.Hash) {
	// 获取交易详情
	tx, _, err := client.TransactionByHash(ctx, txHash)
	if err != nil {
		log.Fatal(err)
	}

	// 获取输入数据 (input data / calldata)
	inputData := tx.Data()

	// 解析和格式化输出
	fmt.Printf("MethodID: 0x%x\n", inputData[:4]) // 前4个字节是 Method ID

	// 将剩余部分格式化
	dataStr := fmt.Sprintf("%x", inputData[4:])

	// 将16进制字符串分割为32字节的块. 每一块是64个字符(Hex)
	var chunks = make([]string, len(dataStr)/64)
	for i := 0; i < len(dataStr); i += 64 {
		end := i + 64
		if end > len(dataStr) {
			end = len(dataStr)
		}
		chunks = append(chunks, dataStr[i:end])
	}
	for i, chunk := range chunks {
		fmt.Printf("[%d]: %s\n", i, chunk)
	}
	// return chunks
}
