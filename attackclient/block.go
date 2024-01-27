package attackclient

import (
	"context"
	"github.com/tsinghua-cel/attacker-service/types"
)

var blockModule = "block"

func (ec *Client) BlockBeforeBroadCast(ctx context.Context) (types.AttackerResponse, error) {
	var result types.AttackerResponse
	err := ec.c.CallContext(ctx, &result, blockModule+"_beforeBroadCast")
	//err := ec.c.CallContext(ctx, &result, blockModule+"_beforeBroadCast", ec.clientInfo())
	if err != nil {
		return result, err
	}
	return result, nil
}

func (ec *Client) BlockAfterBroadCast(ctx context.Context) (types.AttackerResponse, error) {
	var result types.AttackerResponse
	err := ec.c.CallContext(ctx, &result, blockModule+"_afterBroadCast")
	//err := ec.c.CallContext(ctx, &result, blockModule+"_afterBroadCast", ec.clientInfo())
	if err != nil {
		return result, err
	}
	return result, nil
}

func (ec *Client) BlockBeforeSign(ctx context.Context, slot uint64, pubkey string, blockDataBase64 string) (types.AttackerResponse, error) {
	var result types.AttackerResponse
	err := ec.c.CallContext(ctx, &result, blockModule+"_beforeSign", slot, pubkey, blockDataBase64)
	//err := ec.c.CallContext(ctx, &result, blockModule+"_beforeSign", ec.clientInfo(), slot, pubkey, blockDataBase64)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (ec *Client) BlockAfterSign(ctx context.Context, slot uint64, pubkey string, siginedBlockDataBase64 string) (types.AttackerResponse, error) {
	var result types.AttackerResponse
	err := ec.c.CallContext(ctx, &result, blockModule+"_afterSign", slot, pubkey, siginedBlockDataBase64)
	//err := ec.c.CallContext(ctx, &result, blockModule+"_afterSign", ec.clientInfo(), slot, pubkey, siginedBlockDataBase64)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (ec *Client) BlockBeforePropose(ctx context.Context, slot uint64, pubkey string, siginedBlockDataBase64 string) (types.AttackerResponse, error) {
	var result types.AttackerResponse
	err := ec.c.CallContext(ctx, &result, blockModule+"_beforePropose", slot, pubkey, siginedBlockDataBase64)
	//err := ec.c.CallContext(ctx, &result, blockModule+"_beforePropose", ec.clientInfo(), slot, pubkey, siginedBlockDataBase64)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (ec *Client) BlockAfterPropose(ctx context.Context, slot uint64, pubkey string, siginedBlockDataBase64 string) (types.AttackerResponse, error) {
	var result types.AttackerResponse
	err := ec.c.CallContext(ctx, &result, blockModule+"_afterPropose", slot, pubkey, siginedBlockDataBase64)
	//err := ec.c.CallContext(ctx, &result, blockModule+"_afterPropose", ec.clientInfo(), slot, pubkey, siginedBlockDataBase64)
	if err != nil {
		return result, err
	}
	return result, nil
}
