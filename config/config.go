package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type PostgreSQL struct {
	User         string
	Pass         string
	Port         string
	Addr         string
	DatabaseName string
	SslMode      string

	SuperUser     string
	SuperDatabase string
}

type Redis struct {
	Addr string
	Pass string
	User string
}

type RabbitMq struct {
	Addr string
	User string
	Pass string
}

type Config struct {
	Addr    string
	Port    int
	Service string

	Prefix string

	PostgreSQL *PostgreSQL
	Redis      *Redis
	RabbitMq   *RabbitMq
}

var configuration *Config

func loadConfig() {
	if err := godotenv.Load(".env"); err != nil {
		log.Panic(err)
	}

	version := os.Getenv("ADDR")
	if version == "" {
		log.Panic("ADDR")
	}

	portS := os.Getenv("PORT")
	if portS == "" {
		log.Panic("PORT")
	}

	prefix := os.Getenv("PREFIX")
	if prefix == "" {
		log.Panic("PREFIX")
	}

	port, err := strconv.Atoi(portS)
	if err != nil {
		log.Fatalln(err)
	}

	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		log.Panic("SERVICE_NAME")
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		log.Panic("DB_USER")
	}

	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		log.Panic("DB_PASSWORD")
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		log.Panic("DB_PORT")
	}

	dbAddr := os.Getenv("DB_ADDRESS")
	if dbAddr == "" {
		log.Panic("DB_ADDRESS")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Panic("DB_NAME")
	}

	dbSSlmode := os.Getenv("DB_SSLMODE")
	if dbSSlmode == "" {
		log.Panic("DB_SSLMODE")
	}

	dbSuperUser := os.Getenv("DB_SUPERUSER")
	if dbSSlmode == "" {
		log.Panic("DB_SUPERUSER")
	}

	dbSuperDb := os.Getenv("DB_SUPERDB")
	if dbSSlmode == "" {
		log.Panic("DB_SUPERDB")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Panic("REDIS_ADDR")
	}

	// redisUser := os.Getenv("REDIS_USER")
	// if redisUser == "" {
	// 	log.Panic("REDIS_USER")
	// }

	// redisPass := os.Getenv("REDIS_PASS")
	// if redisPass == "" {
	// 	log.Panic("REDIS_PASS")
	// }

	rmqAddr := os.Getenv("RMQ_ADDR")
	if rmqAddr == "" {
		log.Panic("RMQ_ADDR")
	}

	rmqUser := os.Getenv("RMQ_USER")
	if rmqUser == "" {
		log.Panic("RMQ_USER")
	}

	rmqPass := os.Getenv("RMQ_PASS")
	if rmqPass == "" {
		log.Panic("REDIS_PASS")
	}

	configuration = &Config{
		Port:    port,
		Service: serviceName,
		Prefix:  prefix,
		PostgreSQL: &PostgreSQL{
			User:          dbUser,
			Pass:          dbPass,
			Addr:          dbAddr,
			Port:          dbPort,
			DatabaseName:  dbName,
			SslMode:       dbSSlmode,
			SuperUser:     dbSuperUser,
			SuperDatabase: dbSuperDb,
		},
		Redis: &Redis{
			Addr: redisAddr,
			// User: redisUser,
			// Pass: redisPass,
		},
		RabbitMq: &RabbitMq{
			Addr: rmqAddr,
			User: rmqUser,
			Pass: rmqPass,
		},
	}
}

func GetConfig() *Config {
	if configuration == nil {
		loadConfig()
	}
	return configuration
}
