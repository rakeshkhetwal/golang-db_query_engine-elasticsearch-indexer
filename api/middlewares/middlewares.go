package middlewares

import (
	"net/http"
	"time"
	
	log "dashboardapis/api/logger"
)

var standardLogger = log.Logger()

type (
	// holding response details
	responseData struct {
		status int
		size   int
	}

	// custom http.ResponseWriter implementation
	loggingResponseWriter struct {
		http.ResponseWriter 
		responseData        *responseData
	}
)

//getting response size
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b) 
	r.responseData.size += size            
	return size, err
}

//getting status code
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode) 
	r.responseData.status = statusCode      
}

//http logging 
func HttpLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lrw := loggingResponseWriter{
			ResponseWriter: rw, 
			responseData:   responseData,
		}
		rw.Header().Set("Content-Type", "application/json")
		next(&lrw, req)
		duration := time.Since(start)

		//calling standard logger
		standardLogger.HttpLogging(req.RequestURI, req.Method, responseData.status, duration, responseData.size)	
	}
}