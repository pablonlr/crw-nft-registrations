package main

import (
	"github.com/BurntSushi/toml"
)

const CONFIG_PATH = "config.toml"

type Config struct {
	CrwOwnerAddress     string
	CrownRPCAddr        string
	CrownRPCPort        int
	RpcUser             string
	RpcPass             string
	WalletUnlockPass    string
	ClientTimeout       int
	ProtocolID          string
	ProtocolDescription string
	ProtocolSignCode    int
	TokensToMintInRound int
	MintEvery           int
}

func (c *Config) Read() error {
	_, err := toml.DecodeFile(CONFIG_PATH, &c)
	if err != nil {
		return err
	}
	return nil
}
