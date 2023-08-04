package iziswap

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/KyberNetwork/kyberswap-dex-lib/pkg/entity"
)

type PoolsListUpdater struct {
	config *Config
	client *http.Client
}

func NewPoolsListUpdater(
	cfg *Config,
) *PoolsListUpdater {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext,
			// TLSHandshakeTimeout: 10 * time.Second,
		},
		// Timeout: time.Second * 20,
	}
	return &PoolsListUpdater{
		config: cfg,
		client: client,
	}
}

func (d *PoolsListUpdater) GetNewPools(ctx context.Context, metadataBytes []byte) ([]entity.Pool, []byte, error) {

	metadata := Metadata{
		LastCreatedAtTimestamp: 0,
	}

	if metadataBytes != nil || len(metadataBytes) != 0 {
		err := json.Unmarshal(metadataBytes, &metadata)
		if err != nil {
			return nil, metadataBytes, err
		}
	}

	params := &PoolsListQueryParams{
		chainId: d.config.ChainId,
		// todo: for some certain testnet(if exists)
		// we need change version to "v1"
		version:   "v2",
		timeStart: metadata.LastCreatedAtTimestamp,
	}

	// todo: we may implement following request in goroutines
	// for multi pages
	queryResult, err := getPoolsList(d.client, ctx, params, POOL_LIST_LIMIT)

	// despite err may occured during queries
	// but we still return pool info if any exists
	if err != nil && (queryResult != nil || len(queryResult) == 0) {
		return nil, metadataBytes, err
	}

	pools := make([]entity.Pool, 0, len(queryResult))
	latestTimestamp := metadata.LastCreatedAtTimestamp

	for _, p := range queryResult {
		if p.TokenXAddress == emptyString || p.TokenYAddress == emptyString {
			continue
		}
		tokens := make([]*entity.PoolToken, 0, 2)
		reserves := make([]string, 0, 2)

		token0Model := entity.PoolToken{
			Address:   p.TokenXAddress,
			Name:      p.TokenX,
			Symbol:    p.TokenX,
			Decimals:  uint8(p.TokenXDecimals),
			Weight:    defaultTokenWeight,
			Swappable: true,
		}
		tokens = append(tokens, &token0Model)
		reserves = append(reserves, zeroString)

		token1Model := entity.PoolToken{
			Address:   p.TokenYAddress,
			Name:      p.TokenY,
			Symbol:    p.TokenY,
			Decimals:  uint8(p.TokenYDecimals),
			Weight:    defaultTokenWeight,
			Swappable: true,
		}
		tokens = append(tokens, &token1Model)
		reserves = append(reserves, zeroString)

		var swapFee = float64(p.Fee)
		var newPool = entity.Pool{
			Address:      p.Address,
			ReserveUsd:   0,
			AmplifiedTvl: 0,
			SwapFee:      swapFee,
			Exchange:     d.config.DexID,
			Type:         DexTypeiZiSwap,
			Timestamp:    time.Now().Unix(),
			Reserves:     reserves,
			Tokens:       tokens,
		}
		pools = append(pools, newPool)
		if latestTimestamp < p.Timestamp {
			latestTimestamp = p.Timestamp
		}
	}

	newMetadataBytes, err := json.Marshal(Metadata{
		LastCreatedAtTimestamp: latestTimestamp,
	})
	if err != nil {
		return nil, metadataBytes, err
	}

	return pools, newMetadataBytes, nil

}
