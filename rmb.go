package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid-sdk-go/rmb-sdk-go/peer"
	"github.com/threefoldtech/zos/pkg/gridtypes"
	"github.com/urfave/cli"
	"golang.org/x/exp/rand"
)

type rmbCmdArgs map[string]interface{}

func rmbDecorator(action func(c *cli.Context, client *peer.RpcClient) (interface{}, error)) cli.ActionFunc {
	return func(c *cli.Context) error {
		substrate_url := c.String("substrate")
		mnemonics := c.String("mnemonics")
		relay_url := c.String("relay")

		subManager := substrate.NewManager(substrate_url)
		sub, err := subManager.Substrate()
		if err != nil {
			return fmt.Errorf("failed to connect to substrate: %w", err)
		}
		defer sub.Close()

		r := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
		sessionID := fmt.Sprintf("tfgrid-vclient-%d", r.Uint64())

		client, err := peer.NewRpcClient(
			context.Background(),
			mnemonics,
			subManager,
			peer.WithRelay(relay_url),
			peer.WithSession(sessionID),
		)
		if err != nil {
			return fmt.Errorf("failed to create peer client: %w", err)
		}

		res, err := action(c, client)

		if err != nil {
			return err
		}
		fmt.Printf("%v", res)
		return nil

	}
}

func rmbCall(c *cli.Context, client *peer.RpcClient) (interface{}, error) {
	dst := uint32(c.Uint("dst"))
	cmd := c.String("cmd")
	payload := c.String("payload")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var pl interface{}
	if payload != "" {
		if err := json.Unmarshal([]byte(payload), &pl); err != nil {
			return nil, err
		}
	}

	var res interface{}
	if err := client.Call(ctx, dst, cmd, pl, &res); err != nil {
		return nil, err
	}

	b, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

func deploymentChanges(c *cli.Context, client *peer.RpcClient) (interface{}, error) {
	dst := uint32(c.Uint("dst"))
	contractID := c.Uint64("contract_id")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	var changes []gridtypes.Workload
	args := rmbCmdArgs{
		"contract_id": contractID,
	}
	err := client.Call(ctx, dst, "zos.deployment.changes", args, &changes)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment changes after deploy: %w, contractID: %d", err, contractID)
	}
	res, err := json.Marshal(changes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal deployment changes%w", err)
	}
	return string(res), nil
}

func deploymentDeploy(c *cli.Context, client *peer.RpcClient) (interface{}, error) {
	dst := uint32(c.Uint("dst"))
	data := c.String("data")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var dl gridtypes.Deployment
	err := json.Unmarshal([]byte(data), &dl)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal deployment %w", err)
	}

	if err := client.Call(ctx, dst, "zos.deployment.deploy", dl, nil); err != nil {
		return nil, fmt.Errorf("failed to deploy deployment %w", err)
	}

	return nil, nil
}

func deploymentGet(c *cli.Context, client *peer.RpcClient) (interface{}, error) {
	dst := uint32(c.Uint("dst"))
	data := c.String("data")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var args rmbCmdArgs
	err := json.Unmarshal([]byte(data), &args)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data to get deployment %w", err)
	}
	var dl gridtypes.Deployment

	if err := client.Call(ctx, dst, "zos.deployment.get", args, &dl); err != nil {
		return nil, fmt.Errorf("failed to get deployment %w", err)
	}
	json, err := json.Marshal(dl)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal deployment %w", err)
	}

	return string(json), nil
}

func nodeTakenPorts(c *cli.Context, client *peer.RpcClient) (interface{}, error) {
	dst := uint32(c.Uint("dst"))
	var takenPorts []uint16

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := client.Call(ctx, dst, "zos.network.list_wg_ports", nil, &takenPorts); err != nil {
		return nil, fmt.Errorf("failed to get node taken ports %w", err)
	}
	json, err := json.Marshal(takenPorts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal taken ports %w", err)
	}

	return string(json), nil
}

func getNodePublicConfig(c *cli.Context, client *peer.RpcClient) (interface{}, error) {
	dst := uint32(c.Uint("dst"))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var pubConfig struct {
		// Type define if we need to use
		// the Vlan field or the MacVlan
		Type string `json:"type"`
		// Vlan int16     `json:"vlan"`
		// Macvlan net.HardwareAddr

		IPv4 gridtypes.IPNet `json:"ipv4"`
		IPv6 gridtypes.IPNet `json:"ipv6"`

		GW4 net.IP `json:"gw4"`
		GW6 net.IP `json:"gw6"`

		// Domain is the node domain name like gent01.devnet.grid.tf
		// or similar
		Domain string `json:"domain"`
	}

	if err := client.Call(ctx, dst, "zos.network.public_config_get", nil, &pubConfig); err != nil {
		return nil, fmt.Errorf("failed to get node public configuration: %w", err)
	}
	json, err := json.Marshal(pubConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public configuration: %w", err)
	}

	return string(json), nil
}
