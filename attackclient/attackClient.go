package attackclient

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/tsinghua-cel/attacker-service/rpc"
	"math/big"
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

// Delay will delay random seconds time.
func (ec *Client) Delay(ctx context.Context, ts uint) (*big.Int, error) {
	var result hexutil.Big
	err := ec.c.CallContext(ctx, &result, "time_delay", ts)
	if err != nil {
		return nil, err
	}
	return (*big.Int)(&result), err
}

// Delay will delay random seconds time.
func (ec *Client) DelayRandom(ctx context.Context, min, max uint) (*big.Int, error) {
	var result hexutil.Big
	err := ec.c.CallContext(ctx, &result, "time_delayRandom", min, max)
	if err != nil {
		return nil, err
	}
	return (*big.Int)(&result), err
}

func (ec *Client) Echo(ctx context.Context, data string) (string, error) {
	var result string
	err := ec.c.CallContext(ctx, &result, "time_echo", data)
	if err != nil {
		return "", err
	}
	return result, err
}

func (ec *Client) ModifyBlockSlot(ctx context.Context, blockDataBase64 string) (string, error) {
	var result string
	err := ec.c.CallContext(ctx, &result, "block_modifySlot", blockDataBase64)
	if err != nil {
		return "", err
	}
	return result, err
}

func (ec *Client) ModifyAttestSlot(ctx context.Context, blockDataBase64 string) (string, error) {
	var result string
	err := ec.c.CallContext(ctx, &result, "attest_modifySlot", blockDataBase64)
	if err != nil {
		return "", err
	}
	return result, err
}
