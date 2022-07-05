package config

// GeneralConfig will hold the configs
type GeneralConfig struct {
	Alarms        AlarmsConfig
	Notifiers     NotifiersConfig
	InfoTimeOfDay string
}

// AlarmsConfig defines the alarms config
type AlarmsConfig struct {
	NodeRating []NodeRatingAlarmConfig
	NodeNonce  []NodeNonceAlarmConfig
}

// NotifiersConfig defines the implemented notifiers configs
type NotifiersConfig struct {
	Pushover []PushoverNotifier
}

// NodeRatingAlarmConfig the node rating config struct
type NodeRatingAlarmConfig struct {
	Identifier           string
	Threshold            float64
	ApiUrl               string
	PublicKeys           []string
	PollingTimeInSeconds int
}

// NodeNonceAlarmConfig the node's nonce alarm config
type NodeNonceAlarmConfig struct {
	Identifier           string
	ApiUrls              []string
	NonceDifference      int
	PollingTimeInSeconds int
}

// PushoverNotifier pushover's config struct
type PushoverNotifier struct {
	Token string
	User  string
}
