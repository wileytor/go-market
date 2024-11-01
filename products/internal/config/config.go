package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr         string
	DBAddr       string
	MPath        string
	DebugFlag    bool
	RabbitMQHost string
}

const (
	defaultAddr         = ":8081"
	defaultDbDSN        = "postgres://nastya:pgspgs@db:5432/postgres?sslmode=disable"
	defaultMigratePath  = "migrations"
	defaultRabbitMQHost = "rabbitmq"
)

func ReadConfig() Config {
	var addr string
	var dbAddr string
	var migratePath string
	var rabbitMQHost string
	debug := flag.Bool("debug", false, "enable debug logger level")

	flag.StringVar(&addr, "addr", defaultAddr, "Server address") // mani.exe -help
	flag.StringVar(&dbAddr, "db", defaultDbDSN, "database connection addres")
	flag.StringVar(&migratePath, "m", defaultMigratePath, "path to migrations")
	flag.StringVar(&rabbitMQHost, "rabbitMQ", defaultRabbitMQHost, "rabbitMQ host to connect")
	flag.Parse()

	if temp := os.Getenv("SERVER_ADDR"); temp != "" {
		addr = temp
	}
	if temp := os.Getenv("DB_DSN"); temp != "" {
		dbAddr = temp
	}
	if temp := os.Getenv("MIGRATE_PATH"); temp != "" {
		migratePath = temp
	}
	if temp := os.Getenv("RABBITMQ_HOST"); temp != "" {
		rabbitMQHost = temp
	}

	return Config{
		Addr:         addr,
		DBAddr:       dbAddr,
		MPath:        migratePath,
		DebugFlag:    *debug,
		RabbitMQHost: rabbitMQHost,
	}
}
