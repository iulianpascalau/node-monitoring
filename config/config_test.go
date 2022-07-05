package config

import (
	"fmt"
	"testing"

	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
)

func generateMockConfigStruct() GeneralConfig {
	alarmsConfig := AlarmsConfig{
		NodeRating: []NodeRatingAlarmConfig{
			{
				Identifier:           "testnet - rating",
				Threshold:            1,
				ApiUrl:               "http://testnet",
				PublicKeys:           []string{"pk1", "pk2"},
				PollingTimeInSeconds: 5,
			},
			{
				Identifier:           "devnet - rating",
				Threshold:            2,
				ApiUrl:               "http://devnet",
				PublicKeys:           []string{"pk3", "pk4"},
				PollingTimeInSeconds: 6,
			},
		},
		NodeNonce: []NodeNonceAlarmConfig{
			{
				Identifier:           "testnet - nonce",
				ApiUrls:              []string{"http://n1", "http://n2"},
				NonceDifference:      1,
				PollingTimeInSeconds: 2,
			},
			{
				Identifier:           "devnet - nonce",
				ApiUrls:              []string{"http://n2", "http://n3"},
				NonceDifference:      3,
				PollingTimeInSeconds: 4,
			},
		},
	}

	notifiersConfig := NotifiersConfig{
		Pushover: []PushoverNotifier{
			{
				Token: "token1",
				User:  "user1",
			},
			{
				Token: "token2",
				User:  "user2",
			},
		},
	}

	return GeneralConfig{
		Alarms:        alarmsConfig,
		Notifiers:     notifiersConfig,
		InfoTimeOfDay: "11:00:00",
	}
}

func TestMarshalTestConfigs(t *testing.T) {
	cfg := generateMockConfigStruct()

	buff, err := toml.Marshal(cfg)
	assert.Nil(t, err)

	fmt.Println(string(buff))
}

func TestUnmarshal(t *testing.T) {
	expectedConfig := generateMockConfigStruct()

	tomlData := `
InfoTimeOfDay = "11:00:00"

[Alarms]
  [[Alarms.NodeNonce]]
    ApiUrls = ["http://n1", "http://n2"]
    Identifier = "testnet - nonce"
    NonceDifference = 1
    PollingTimeInSeconds = 2

  [[Alarms.NodeNonce]]
    ApiUrls = ["http://n2", "http://n3"]
    Identifier = "devnet - nonce"
    NonceDifference = 3
    PollingTimeInSeconds = 4

  [[Alarms.NodeRating]]
    ApiUrl = "http://testnet"
    Identifier = "testnet - rating"
    PollingTimeInSeconds = 5
    PublicKeys = ["pk1", "pk2"]
    Threshold = 1.0

  [[Alarms.NodeRating]]
    ApiUrl = "http://devnet"
    Identifier = "devnet - rating"
    PollingTimeInSeconds = 6
    PublicKeys = ["pk3", "pk4"]
    Threshold = 2.0

[Notifiers]
  [[Notifiers.Pushover]]
    Token = "token1"
    User = "user1"

  [[Notifiers.Pushover]]
    Token = "token2"
    User = "user2"
`

	result := GeneralConfig{}
	err := toml.Unmarshal([]byte(tomlData), &result)
	assert.Nil(t, err)
	assert.Equal(t, expectedConfig, result)
}
