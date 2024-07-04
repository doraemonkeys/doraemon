package ethereum

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func QueryAccountETHBalance(ctx context.Context, client *ethclient.Client, account common.Address) (balance *big.Float, err error) {
	balanceWei, err := client.BalanceAt(ctx, account, nil)
	if err != nil {
		return nil, err
	}
	// Convert Wei to ETH
	balance = new(big.Float).Quo(new(big.Float).SetInt(balanceWei), new(big.Float).SetFloat64(1e18))
	return balance, nil
}
