package clients

import (
	"github.com/rs/zerolog"
	"github.com/scmhub/ibapi"
)

type IBAPIEClient struct {
	*ibapi.EClient
	logger zerolog.Logger
}

func NewIBAPIEClient(logger zerolog.Logger) *IBAPIEClient {
	return &IBAPIEClient{
		EClient: ibapi.NewEClient(nil),
		logger:  logger,
	}
}
