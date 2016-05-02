package main

import (
	"os"
	"log"
	"time"
	"github.com/sandstorm/mailer-daemon/router"
	"github.com/sandstorm/mailer-daemon/recipientsRepository"
	"github.com/garyburd/redigo/redis"
	"github.com/sandstorm/mailer-daemon/worker"
	"net/smtp"
	"io"
)

func main() {
	log.Println("Starting mailer...")

	logFile := configureLogger()
	defer logFile.Close()

	serverConfiguration, emailGateway := getConfiguration()

	go startSlave(serverConfiguration.RedisConfiguration, emailGateway)
	startMaster(serverConfiguration)
}

func configureLogger() *os.File {
	logFile, err := os.OpenFile("mailer.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil { log.Fatal("failed to open log file", err) }

	log.SetOutput(io.MultiWriter(os.Stderr, logFile))
	return logFile
}

func getConfiguration() (router.ServerConfiguration, worker.EmailGateway) {
	redisUrl := os.Getenv("REDIS_URL")
	if len(redisUrl) <= 0 { redisUrl = "localhost:6379" }

	emailGateway := createEmailGateway()
	return router.ServerConfiguration{
		RedisConfiguration: router.RedisConfiguration{
			RedisUrl: redisUrl,
			Verbosity: os.Getenv("VERBOSITY"),
		},
		AuthToken: os.Getenv("AUTH_TOKEN"),
		EmailGateway: emailGateway.Description(),
	}, emailGateway
}

func createEmailGateway() worker.EmailGateway {
	mandrillApiKey := os.Getenv("MANDRILL_API_KEY")
	if len(mandrillApiKey) > 0 {
		return createMandrillGateway(mandrillApiKey)
	}

	smtpUrl := os.Getenv("SMTP_URL")
	if len(smtpUrl) > 0 {
		return createSmtpGateway(smtpUrl)
	}
	return createSmtpGateway("localhost:25")
}

func createSmtpGateway(url string) worker.EmailGateway {
	log.Println("\tusing SMTP-Gateway at " + url)
	return &worker.SmtpEmailGateway{
		SmtpUrl: url,
		SmtpAuth: smtp.PlainAuth("", "user@example.com", "password", "localhost"),
	}
}

func createMandrillGateway(apiKey string) worker.EmailGateway {
	log.Println("\tusing Mandrill-Gateway")
	return &worker.MandrillGateway{
		ApiKey: apiKey,
	}
}

func startMaster(config router.ServerConfiguration) {
	log.Println("Starting master...")

	server := router.Server{
		AuthToken: config.AuthToken,
		Repository: createRecipientRepository(config.RedisConfiguration),
		ServerConfiguration: config,
	}

	err := server.Listen()
	if err != nil { log.Fatal(err) }
}

func createRecipientRepository(config router.RedisConfiguration) (repository recipientsRepository.Repository) {
	repository = &recipientsRepository.RedisRepository{
		Pool: createRedisPool(config.RedisUrl),
	}

	if len(config.Verbosity) > 0 {
		repository = &recipientsRepository.LoggingDummyRepository{
			Repository: repository,
		}
	}

	return repository
}

func startSlave(redisConfiguration router.RedisConfiguration, emailGateway worker.EmailGateway) {
	log.Println("Starting slave...")

	emailSender := worker.EmailSender{
		Gateway: emailGateway,
		Repository: createRecipientRepository(redisConfiguration),
	}
	emailSender.Start()
}

func createRedisPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxActive: 10,
		Wait: true,
		MaxIdle: 3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			connection, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			//if _, err := connection.Do("AUTH", password); err != nil {
			//	connection.Close()
			//	return nil, err
			//}
			return connection, err
		},
		TestOnBorrow: func(connection redis.Conn, t time.Time) error {
			_, err := connection.Do("PING")
			return err
		},
	}
}
