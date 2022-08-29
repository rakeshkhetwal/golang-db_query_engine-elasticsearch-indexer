package controllers

import ( "dashboardapis/api/middlewares"
	    )  

//routes intialize
func (s *Server) initializeRoutes() {
	// Home Route
	s.Router.HandleFunc("/", middlewares.HttpLogging(s.Home)).Methods("GET")
	
	// SQL query executer
	s.Router.HandleFunc("/query/", middlewares.HttpLogging(s.getQueryResult)).Methods("GET")
	
	//elasticsearch api 
	s.Router.HandleFunc("/elastic/save/", middlewares.HttpLogging(s.postQueryResult)).Methods("POST")
	
	//logging routes initialization
	standardLogger.SuccessMessage("Routes successfully initialized", "")
}
