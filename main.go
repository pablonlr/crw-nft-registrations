package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"crypto/sha256"

	crown "github.com/pablonlr/go-rpc-crownd"
)

func hash1(value string) string {
	hasher := sha256.New()
	hasher.Write([]byte(value))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func main() {
	configuration := &Config{}
	err := configuration.Read()
	if err != nil {
		panic(err)
	}
	client, err := crown.NewClient(configuration.CrownRPCAddr, configuration.CrownRPCPort, configuration.RpcUser, configuration.RpcPass, configuration.ClientTimeout)
	if err != nil {
		panic(err)
	}

	//log events to external file
	logFile, err := os.OpenFile("log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer logFile.Close()
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(logFile)

	if !protocolExist(configuration.ProtocolID, client) {
		err = registerProtocol(configuration, client)
		if err != nil {
			log.Fatalln("error in protocol registration : ", err)
		}
	}

	err = mintNTokens(configuration, client, true)
	if err != nil {
		log.Println("error minting token on init: ", err.Error())
	}

	//init ticker worker
	ticker := time.NewTicker(time.Duration(configuration.MintEvery) * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:

				log.Println("New NFT Minting Round", t)
				//unlockWallet

				err := mintNTokens(configuration, client, false)
				if err != nil {
					log.Println("error minting token: ", err.Error(), t)
				}
			}
		}
	}()
	select {}

}

func mintNTokens(conf *Config, client *crown.Client, returnOnErr bool) error {
	n := conf.TokensToMintInRound
	protocol := conf.ProtocolID
	addr := conf.CrwOwnerAddress
	client.Unlock(conf.WalletUnlockPass, conf.TokensToMintInRound*5)
	for i := 0; i < n; i++ {
		st := strconv.Itoa(int(time.Now().Unix()) + i)
		id := hash1(st)
		meta := fmt.Sprintf("test token %s", id)
		result, err := client.RegisterNFToken(protocol, id, addr, addr, meta)
		if err != nil {
			if returnOnErr {
				return err
			}
			log.Println("error minting token: ", err.Error())
			continue
		}
		log.Println("new registration: ", result)
	}
	return nil
}

// register a new nft protocol with a given configuration
func registerProtocol(configuration *Config, client *crown.Client) error {
	//try to register the new protocol
	client.Unlock(configuration.WalletUnlockPass, configuration.TokensToMintInRound*5)
	_, err := client.RegisterNFTProtocol(configuration.ProtocolID, configuration.ProtocolDescription, configuration.CrwOwnerAddress, 2, "application/json", configuration.ProtocolDescription, false, true, 255)
	if err != nil {
		return err
	}
	return nil
}

// check if the nft protocol is already registered
func protocolExist(id string, client *crown.Client) bool {
	proto, err := client.GetNFTProtocol(id)
	if err != nil || len(proto.RegistrationTxHash) < 1 {
		return false
	}
	return true
}
