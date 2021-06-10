package ethereum

type Config struct {
	Endpoint              string   `mapstructure:"endpoint"`
	PrivateKey            string   `mapstructure:"private-key"`
	DescendantsUntilFinal byte     `mapstructure:"descendants-until-final"`
	Channels              Channels `mapstructure:"channels"`
	LightClientBridge     string   `mapstructure:"lightclientbridge"`
}

type Channels struct {
	Basic        Channel `mapstructure:"basic"`
	Incentivized Channel `mapstructure:"incentivized"`
}

type Channel struct {
	Inbound  string `mapstructure:"inbound"`
	Outbound string `mapstructure:"outbound"`
}
