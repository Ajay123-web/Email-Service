package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "8000"

func main() {
	//connect to database
	db := initDB()

	//db.Ping()

	//create sessions

	session := initSession()

	//create loggers
	infoLog := log.New(os.Stdout , "INFO\t" , log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout , "ERROR\t" , log.Ldate|log.Ltime|log.Lshortfile)

	//create channels

	//create waitgroups

	wg := sync.WaitGroup{}

	//set config

	app := Config{
		DB: db,
		Wait: &wg,
		Session: session,
		InfoLog: infoLog,
		ErrorLog: errorLog,
	}

	//set up mail

	//listen for web connection
	app.serve()
}

func (app *Config) serve() {
	srv := &http.Server{
		Addr: webPort,
		Handler: app.routes(),
	}

	app.InfoLog.Println("Starting server...")
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic("Server not starting...")
	}
}

func initDB() *sql.DB {
	conn := connectToDB()

	if conn == nil {
		log.Panic("Cant connect to database")
	}
	return conn
}

func connectToDB() *sql.DB {
	cnt := 0
	dsn := os.Getenv("DSN")
	for {
		conn , err := openDB(dsn)
		if err != nil {
			log.Print("Postgres not ready yet...")
		} else {
			log.Print("Connected to database")
			return conn
		}
		if cnt > 10 {
			return nil
		}
		time.Sleep(1 * time.Second)
		cnt += 1
	}
}

func openDB(dsn string) (*sql.DB , error) {
	db , err := sql.Open("pgx" , dsn)
	if err != nil {
		return nil , err
	}

	err = db.Ping()
	if err != nil {
		return nil , err
	} 
	return db , nil

}

func initSession() *scs.SessionManager {
	session := scs.New()
	session.Store = redisstore.New(initRedis())
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true
	return session
}

func initRedis() *redis.Pool {
	return &redis.Pool {
		MaxIdle: 10,
		Dial: func() (redis.Conn , error) {
			return redis.Dial("tcp" , os.Getenv("REDIS"))
		},
	}
}