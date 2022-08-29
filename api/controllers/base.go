package controllers

import (
	"fmt"
	"net/http"
    "database/sql"
    "os"
	"errors"

	"github.com/joho/godotenv"
	log "dashboardapis/api/logger"
	"github.com/gorilla/mux"
	
)

type Server struct {
	Router *mux.Router
}	

var server = Server{}
var standardLogger = log.Logger()
//null error handler message
var nilerr = errors.New("")

//Env var load from .env file
func envVarLoad() {
	var err error
	err = godotenv.Load(".env")

	if err != nil {
		standardLogger.FatalErrorMessage(err,"")
	} else {
		standardLogger.SuccessMessage("Environment values loaded", "")
	} 
}

//Initialize the DB
func Initialize(dbDriver, dbName string) (*sql.DB, error) {
	// loading Env variables
	envVarLoad()

	if dbDriver == "mysql" {
		DBURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST") , os.Getenv("DB_PORT"), dbName)
		db, err := sql.Open(os.Getenv("DB_DRIVER"), DBURL)
		if err != nil {
			standardLogger.ErrorMessage(err,"")
		} else {
			standardLogger.SuccessMessage("Connected to the database :", dbDriver)
		}
		return db, nil

	} else if dbDriver == "postgres" {
		DBURL := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", os.Getenv("POSTGRESS_DB_HOST"), os.Getenv("POSTGRESS_DB_PORT"), os.Getenv("POSTGRESS_DB_USER"), dbName, os.Getenv("POSTGRESS_DB_PASSWORD"))
		db, err := sql.Open(os.Getenv("DB_DRIVER"), DBURL)
		if err != nil {
			standardLogger.ErrorMessage(err, "")
		} else {
			standardLogger.SuccessMessage("Connected to the database :", dbDriver)
		}
		return db, nil
	} else {		
		standardLogger.ErrorMessage(nilerr, "Incorrect DB driver provided")
	}

	return nil, nil
}

func Run() {
    // Router  initialize
    server.Router = mux.NewRouter()
	server.initializeRoutes()
	// Server Listen
	server.RunServer(":8084")

}

func (server *Server) RunServer(addr string) {
	standardLogger.SuccessMessage("Starting server at", "8084")
	standardLogger.FatalErrorMessage(http.ListenAndServe(addr, server.Router),"")
}
