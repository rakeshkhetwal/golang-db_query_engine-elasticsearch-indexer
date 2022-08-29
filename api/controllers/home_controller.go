package controllers

import (
	"net/http"
	
	"dashboardapis/api/responses"
)

func (server *Server) Home(w http.ResponseWriter, r *http.Request) {
	responses.JSON(w, http.StatusOK, "Welcome to Database Query engine and Elastic Search bulk posting engine")
}
