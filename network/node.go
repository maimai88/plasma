package network

import (
	// "github.com/icstglobal/plasma/core"
	// "github.com/ethereum/go-ethereum/node"
	"encoding/json"
	// "fmt"
	"os"

	"github.com/icstglobal/plasma/plasma"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//Node is a P2P network member, it can acts both as client and server
type Node struct {
	rpc  *RPCServer
	http *HTTPServer
	p2p  *P2PLocalHost
}

type CxData struct {
	Abi interface{} `json:"abi"`
	Bin string      `json:"bytecode"`
}

//Start the node
func (n *Node) Start() (err error) {
	datadir := viper.GetString("plasma.datadir")
	networkId := viper.GetInt64("plasma.networkId")
	chainUrl := viper.GetString("rootchain.url")
	cxAddr := viper.GetString("rootchain.cxAddr")
	cxFileName := viper.GetString("rootchain.cxFileName")
	operPwd := viper.GetString("rootchain.operPwd")
	keystorePath := viper.GetString("rootchain.keystorePath")

	cfg := &plasma.DefaultConfig
	cfg.DataDir = datadir
	cfg.NetworkId = uint64(networkId)
	cfg.ChainUrl = chainUrl
	cfg.CxAddr = cxAddr
	cfg.OperPwd = operPwd
	cfg.KeystorePath = keystorePath
	// read abi from json
	cxData := new(CxData)
	file, err := os.Open(cxFileName)
	defer file.Close()
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cxData)
	if err != nil {
		return err
	}
	abiStr, err := json.Marshal(cxData.Abi)
	if err != nil {
		return err
	}

	cfg.CxAbi = string(abiStr)
	// log.Debugf("abi: %v %v\n", cfg.CxAbi, reflect.TypeOf(cxData.Abi))

	plasma, err := plasma.New(cfg)
	n.rpc, err = startRPC(plasma)
	if err != nil {
		return err
	}
	log.Debug("plasma:", plasma)

	n.http, err = startHTTP(plasma)
	if err != nil {
		return err
	}

	n.p2p, err = startP2P(plasma)
	if err != nil {
		return err
	}

	return nil
}

func startRPC(plasma *plasma.Plasma) (*RPCServer, error) {
	proto := viper.GetString("rpcserver.proto")
	port := viper.GetInt("rpcserver.port")
	log.Info("try to start rpc server")
	rpcCfg := RPCConfig{
		Proto:   proto,
		Port:    port,
		Methods: calls(),
	}
	rpc, err := ServeRPC(rpcCfg)
	if err != nil {
		return nil, errors.Annotate(err, "failed to start rpc server")
	}
	log.WithFields(log.Fields{"proto": proto, "port": port}).Info("rpc server is running")

	return rpc, nil
}

func startHTTP(plasma *plasma.Plasma) (*HTTPServer, error) {
	log.Info("try to start http server")
	httpPort := viper.GetInt("httpserver.port")

	http := &HTTPServer{
		Port:   httpPort,
		Plasma: plasma,
	}
	if err := http.Start(); err != nil {
		return nil, errors.Annotate(err, "failed to start http server")
	}
	log.WithField("port", httpPort).Info("http server is running")

	return http, nil
}

func calls() map[string]interface{} {
	calls := make(map[string]interface{})
	//TODO: use real calls
	var fakeService RPCService
	calls["fake"] = fakeService
	return calls
}
