package sdk

import (
	"net/http"
	"strings"

	"github.com/alecsavvy/mojave/gen/api/v1/v1connect"
)

type SonataSDK struct {
	Chain       v1connect.ChainClient
	Storage     v1connect.StorageClient
	System      v1connect.SystemClient
	P2P         v1connect.P2PClient
	DDEX        v1connect.DDEXClient
	Composition v1connect.CompositionClient
	Account     v1connect.AccountClient
	Validator   v1connect.ValidatorClient
}

func NewSonataSDK(url string) *SonataSDK {
	url = ensureURLProtocol(url)
	client := http.DefaultClient
	return &SonataSDK{
		Chain:       v1connect.NewChainClient(client, url),
		Storage:     v1connect.NewStorageClient(client, url),
		System:      v1connect.NewSystemClient(client, url),
		P2P:         v1connect.NewP2PClient(client, url),
		DDEX:        v1connect.NewDDEXClient(client, url),
		Composition: v1connect.NewCompositionClient(client, url),
		Account:     v1connect.NewAccountClient(client, url),
		Validator:   v1connect.NewValidatorClient(client, url),
	}
}

func ensureURLProtocol(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "https://" + url
	}
	return url
}
