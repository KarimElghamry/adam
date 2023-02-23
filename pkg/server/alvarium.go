package server

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"sync"

	"github.com/KarimElghamry/alvarium-sdk-go/pkg"
	"github.com/KarimElghamry/alvarium-sdk-go/pkg/config"
	"github.com/KarimElghamry/alvarium-sdk-go/pkg/factories"
	"github.com/KarimElghamry/alvarium-sdk-go/pkg/interfaces"
	logConfig "github.com/project-alvarium/provider-logging/pkg/config"
	logFactory "github.com/project-alvarium/provider-logging/pkg/factories"
	"github.com/project-alvarium/provider-logging/pkg/logging"
)

var sdk interfaces.Sdk

type ApplicationConfig struct {
	Sdk     config.SdkInfo        `json:"sdk,omitempty"`
	Logging logConfig.LoggingInfo `json:"logging,omitempty"`
}

func (a ApplicationConfig) AsString() string {
	b, _ := json.Marshal(a)
	return string(b)
}

func getAlvariumSdk() (interfaces.Sdk, error) {
	if sdk != nil {
		return sdk, nil
	}

	// Load config
	var configPath string
	flag.StringVar(&configPath,
		"cfg",
		"./res/alvarium-config.json",
		"Path to JSON configuration file.")
	flag.Parse()

	jsonFile, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()

	var cfg ApplicationConfig
	configBytes, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(configBytes, &cfg)
	if err != nil {
		return nil, err
	}

	logger := logFactory.NewLogger(cfg.Logging)
	logger.Write(logging.DebugLevel, "config loaded successfully")
	logger.Write(logging.DebugLevel, cfg.AsString())

	// init annotators
	var annotators []interfaces.Annotator
	for _, t := range cfg.Sdk.Annotators {
		instance, err := factories.NewAnnotator(t, cfg.Sdk)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
		annotators = append(annotators, instance)
	}

	sdk = pkg.NewSdk(annotators, cfg.Sdk, logger)
	var wg sync.WaitGroup
	sdk.BootstrapHandler(context.Background(), &wg)
	return sdk, nil
}
