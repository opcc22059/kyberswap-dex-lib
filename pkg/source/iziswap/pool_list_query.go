package iziswap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var urlRoot = "https://api.izumi.finance/api/v1/izi_swap/meta_record/"

// params = "chain_id=324&type=10&version=v2&time_start=2023-06-02T13:53:13&page=1&page_size=10&order_by=time"

type PoolsListQueryParams struct {
	chainId int
	// v1 or v2
	version string
	// timestamp in second
	timeStart int
}

type PoolInfo struct {
	Fee            int    `json:"fee"`
	TokenX         string `json:"tokenX"`
	TokenY         string `json:"tokenY"`
	Address        string `json:"address"`
	Timestamp      int    `json:"timestamp"`
	TokenXAddress  string `json:"tokenX_address"`
	TokenYAddress  string `json:"tokenY_address"`
	TokenXDecimals int    `json:"tokenX_decimals"`
	TokenYDecimals int    `json:"tokenY_decimals"`
	Version        string `json:"version"`
}

type PoolsListQueryResponse struct {
	Data  *[]PoolInfo `json:"data,omitempty"`
	Total int         `json:"total"`
}

func getPoolsList(
	client *http.Client,
	ctx context.Context,
	params *PoolsListQueryParams,
	limit int,
) ([]PoolInfo, error) {

	url := fmt.Sprintf(
		"%s?chain_id=%d&type=10&version=%s&page_size=%d&order_by=time&format=json",
		urlRoot,
		params.chainId,
		params.version,
		POOL_LIST_PAGE_SIZE,
	)
	capacity := limit
	if capacity < 0 {
		capacity = POOL_LIST_LIMIT
	}
	first := true
	page := 1
	var result []PoolInfo = nil
	cnt := 0
	for {
		currentUrl := fmt.Sprintf(
			"%s&page=%d",
			url,
			page,
		)
		req, err := http.NewRequest("GET", currentUrl, nil)
		if err != nil {
			return result, err
		}
		req = req.WithContext(ctx)
		resp, err := client.Do(req)
		if err != nil {
			return result, err
		}

		data, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return result, err
		}

		if err != nil {
			return result, err
		}
		var response PoolsListQueryResponse

		if err = json.Unmarshal(data, &response); err != nil {
			return result, err
		}

		if first {
			if limit > response.Total {
				limit = response.Total
			}
			result = make([]PoolInfo, 0, limit)
			first = false
		}
		currentSize := 0
		if response.Data != nil {
			currentSize = len(*response.Data)
		}
		if currentSize > 0 {
			result = append(result, *response.Data...)
		}
		cnt += currentSize
		if cnt >= limit || currentSize < POOL_LIST_PAGE_SIZE {
			break
		}
		page += 1
	}
	return result, nil
}
