package controllers

import (
	"net/http"
	"regexp"

     _ "github.com/go-sql-driver/mysql"
	"github.com/xwb1989/sqlparser"

	"dashboardapis/api/responses"
	"dashboardapis/api/error_handler"
	"github.com/rakeshkhetwal/sqltojson"
)

// Allowed operation for query to execute
func operationValidator(query string)bool {
	var match_pattern = regexp.MustCompile(`^select|SELECT`)
	valid_operation:=match_pattern.MatchString(query)
    if valid_operation{
		return true
	}
	return false
} 

//check if query is syntactically correct or not
func queryValidator(query string) (bool, error){
	_, err := sqlparser.Parse(query)
    if err != nil {
        return false, err
    } else {
		valid_operation:=operationValidator(query)
		return valid_operation, nil 
	}
    return true, nil
}

//Check if params for null value
func paramsNullHandler(dbDriver, dbName, query string) bool{
	if dbDriver != "" && dbName != "" && query != "" {
		return true
	}
	return false
}

//Db driver validator
func dbDriverValidator(dbDriver string) bool{
	if dbDriver == "mysql" || dbDriver == "postgres" {
		return true
	}
	return false
}

//DB Query Executer
func (server *Server) getQueryResult(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	dbDriver := params.Get("dbDriver")
	dbName := params.Get("dbName")
	query := params.Get("query")
    
	//Db driver validator
	dbDriver_Validator:=dbDriverValidator(dbDriver)
	if dbDriver_Validator == false{
		standardLogger.ErrorMessage(nilerr,"Incorrect DB driver provided")
		db_driver_error := error_handler.IncorrectDbDriver()
		responses.ERROR(w, http.StatusBadRequest, db_driver_error)
		return
	}

	//param validation for none value
	param_validation:=paramsNullHandler(dbDriver, dbName, query)
	if param_validation == false {
		standardLogger.ErrorMessage(nilerr,"Incomplete parameters passed")
		param_validation_error := error_handler.ParamsNullHandler()
		responses.ERROR(w, http.StatusBadRequest, param_validation_error)
		return
	}

	validated_query, err:= queryValidator(query)
	if err != nil || validated_query == false {
		// for unauthorized query
		if validated_query == false && err == nil {
			standardLogger.ErrorMessage(nilerr,"Unauthorized operation, please use select operation only")
			unauthorised_error := error_handler.UnauthorizedHandler()
			responses.ERROR(w, http.StatusUnauthorized, unauthorised_error)
			return
		} else {
			//syntax error response
			standardLogger.ErrorMessage(err,"")
			responses.ERROR(w, http.StatusBadRequest, err)
			return
		}
		return
	}

	//db initialize
	db, err := Initialize(dbDriver, dbName)
	//sql to json conversion
	err,queryData:=sqltojson.SqlToJson(db,query)

	if err != nil {
		standardLogger.ErrorMessage(err,"")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, queryData)
}
