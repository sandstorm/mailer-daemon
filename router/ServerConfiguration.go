package router

type ServerConfiguration struct {
	RedisConfiguration RedisConfiguration
	AuthToken string
	EmailGateway string
}

type RedisConfiguration struct {
	RedisUrl string
	Verbosity string
}