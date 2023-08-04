package iziswap

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/KyberNetwork/ethrpc"
	"github.com/KyberNetwork/kyberswap-dex-lib/pkg/source/pool"
	"github.com/ethereum/go-ethereum/common"
)

func TestPoolsListTrackerAndSimulator(t *testing.T) {
	cfg := &Config{
		DexID:      "",
		ChainId:    5000,
		PointRange: 2000,
	}
	updater := NewPoolsListUpdater(cfg)
	ctx, _ := context.WithCancel(context.Background())
	pools, _, _ := updater.GetNewPools(ctx, nil)
	fmt.Println(len(pools))

	poolRaw := pools[2]

	rpcUrl := "https://rpc.mantle.xyz"
	client := ethrpc.New(rpcUrl)
	client.SetMulticallContract(common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"))

	poolTracker, _ := NewPoolTracker(cfg, client)
	p, _ := poolTracker.GetNewPoolState(ctx, poolRaw)

	poolSimulator, _ := NewPoolSimulator(p)
	fmt.Println(poolSimulator.PoolInfo)

	fmt.Println(p.Tokens[0].Symbol, p.Tokens[0].Decimals, p.Tokens[0].Address)
	fmt.Println(p.Tokens[1].Symbol, p.Tokens[1].Decimals, p.Tokens[1].Address)

	// usdt to eth
	// input amount is 1000.0 usdt
	inputAmount0, _ := new(big.Int).SetString("1000000000", 10)
	tokenAmount0 := pool.TokenAmount{
		Token:     p.Tokens[0].Address,
		Amount:    inputAmount0,
		AmountUsd: 0,
	}
	res, err := poolSimulator.CalcAmountOut(tokenAmount0, p.Tokens[1].Address)
	if err != nil {
		fmt.Println("err: ", err)
	} else {
		fmt.Println(res.TokenAmountOut.Amount)
	}

	// eth to usdt
	// input amount is 0.5 eth
	inputAmount1, _ := new(big.Int).SetString("500000000000000000", 10)
	tokenAmount1 := pool.TokenAmount{
		Token:     p.Tokens[1].Address,
		Amount:    inputAmount1,
		AmountUsd: 0,
	}
	res1, err := poolSimulator.CalcAmountOut(tokenAmount1, p.Tokens[0].Address)
	if err != nil {
		fmt.Println("err: ", err)
	} else {
		fmt.Println(res1.TokenAmountOut.Amount)
	}
}
