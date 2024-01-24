package attackclient

import (
	"context"
	"github.com/tsinghua-cel/attacker-service/rpc"
	"github.com/tsinghua-cel/attacker-service/types"
)

// Client defines typed wrappers for the Ethereum RPC API.
type Client struct {
	c *rpc.Client
}

// Dial connects a client to the given URL.
func Dial(rawurl string) (*Client, error) {
	return DialContext(context.Background(), rawurl)
}

// DialContext connects a client to the given URL with context.
func DialContext(ctx context.Context, rawurl string) (*Client, error) {
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	return NewClient(c), nil
}

// NewClient creates a client that uses the given RPC client.
func NewClient(c *rpc.Client) *Client {
	return &Client{c}
}

// Close closes the underlying RPC connection.
func (ec *Client) Close() {
	ec.c.Close()
}

// Client gets the underlying RPC client.
func (ec *Client) Client() *rpc.Client {
	return ec.c
}

func (ec *Client) BlockBroadCastDelay(ctx context.Context) (types.AttackerResponse, error) {
	var result types.AttackerResponse
	err := ec.c.CallContext(ctx, &result, "block_broadCastDelay")
	if err != nil {
		return result, err
	}
	return result, nil
}

func (ec *Client) BlockModify(ctx context.Context, slot int64, pubkey string, blockDataBase64 string) (types.AttackerResponse, error) {
	var result types.AttackerResponse
	err := ec.c.CallContext(ctx, &result, "block_modifyBlock", slot, pubkey, blockDataBase64)
	if err != nil {
		return result, err
	}
	return result, err
}

func (ec *Client) AttestBroadCastDelay(ctx context.Context) (types.AttackerResponse, error) {
	var result types.AttackerResponse
	err := ec.c.CallContext(ctx, &result, "attest_broadCastDelay")
	if err != nil {
		return result, err
	}
	return result, nil
}

func (ec *Client) AttestModify(ctx context.Context, slot int64, pubkey string, attestDataBase64 string) (types.AttackerResponse, error) {
	var result types.AttackerResponse
	err := ec.c.CallContext(ctx, &result, "attest_modifyAttest", slot, pubkey, attestDataBase64)
	if err != nil {
		return result, err
	}
	return result, err
}

// Delay will delay random seconds time.
func (ec *Client) Delay(ctx context.Context, ts uint) (types.AttackerResponse, error) {
	var result types.AttackerResponse
	err := ec.c.CallContext(ctx, &result, "time_delay", ts)
	if err != nil {
		return result, err
	}
	return result, err
}

// Delay will delay random seconds time.
func (ec *Client) DelayRandom(ctx context.Context, min, max uint) (types.AttackerResponse, error) {
	var result types.AttackerResponse
	err := ec.c.CallContext(ctx, &result, "time_delayRandom", min, max)
	if err != nil {
		return result, err
	}
	return result, err
}

func (ec *Client) Echo(ctx context.Context, data string) (types.AttackerResponse, error) {
	var result types.AttackerResponse
	err := ec.c.CallContext(ctx, &result, "time_echo", data)
	if err != nil {
		return result, err
	}
	return result, err
}
