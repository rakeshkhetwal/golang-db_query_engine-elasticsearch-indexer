package controllers

import (
	"fmt"
	"net/http"
	"os"
	"context"
	"strconv"
	"strings"
	"io/ioutil"
	"net/url"
	"errors"
	"flag"
	"time"
	"runtime"
	"sync/atomic"
	"encoding/json"
	
	"dashboardapis/api/responses"

	// "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/cenkalti/backoff/v4"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/dustin/go-humanize"

	
	
)

var (
	r  map[string]interface{}
	indexName  string
	numWorkers int
	flushBytes int
	countSuccessful uint64
	res *esapi.Response
	err error
)

//initializing default bulk indexer config
func init() {
	flag.StringVar(&indexName, "index", "", "Index name")
	flag.IntVar(&numWorkers, "workers", runtime.NumCPU(), "Number of indexer workers")
	flag.IntVar(&flushBytes, "flush", 5e+6, "Flush threshold in bytes")
	flag.Parse()
}

//Initializing the Elastic search client
func getESClient() (*elasticsearch.Client, error) {
	retryBackoff := backoff.NewExponentialBackOff()
	cfg := elasticsearch.Config{
		Addresses: []string{os.Getenv("ELASTICSEARCH_HOST")},
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retryBackoff.Reset()
			}
			return retryBackoff.NextBackOff()
		},
		Username:      os.Getenv("USERNAME"),
		Password:      os.Getenv("PASSWORD"),
		MaxRetries:    5,
		RetryOnStatus: []int{502, 503, 504, 429},
	}

	es, err := elasticsearch.NewClient(cfg)

	if err != nil {
		standardLogger.FatalErrorMessage(err, "Error intializing Elastic Search client")
		return nil, err
	}

	standardLogger.SuccessMessage("Successfully initialized Elastic search client", "")

	// Getting ES client info
	res, err := es.Info()

	if err != nil {
		standardLogger.ErrorMessage(err,"Error getting response")
	}
	defer res.Body.Close()

	// Check response status
	if res.IsError() {
		res_error := errors.New(res.String())
		standardLogger.ErrorMessage(res_error,"Error getting response")
	}

	// Deserialize the response into a map.
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		standardLogger.ErrorMessage(err,"Error parsing the response body")
	}

	// Print client and server version numbers.
	standardLogger.InfoMessage("Elastic Search Client Version : ",elasticsearch.Version)
	standardLogger.InfoMessage("Elastic Search Server Version : ",fmt.Sprint(r["version"].(map[string]interface{})["number"]))

	return es, nil
}

//Extracting the query result from database.
func extractQueryResult(dbDriver, dbName, query string) ([]string, error){
	var data []string
	//creating a map from bytes	
	var objmap []map[string]interface{}

	standardLogger.InfoMessage("Performing Http Get .... ", "")
    query_response, err := http.Get(`http://localhost:8084/query/?dbDriver=`+ dbDriver + `&dbName=` + dbName + `&query=` + url.QueryEscape(query))

	if err != nil {
		standardLogger.ErrorMessage(err,"Error fetching the GET response :")
		return nil, err
    }

    defer query_response.Body.Close()
	//get response in bytes
    query_response_body_bytes, _ := ioutil.ReadAll(query_response.Body)
	

	if query_response.StatusCode != 200 {
		error_respone := errors.New(string(query_response_body_bytes))
		status := strconv.Itoa(query_response.StatusCode)
		data=append(data, status)
		return data, error_respone
	}
	
	//converts json to string
	json.Unmarshal(query_response_body_bytes, &objmap)

	for  _, value := range objmap {
		value, err := json.Marshal(value)
		if err != nil {
			standardLogger.ErrorMessage(err,"Error marshalling data :")
		} 
		data=append(data, string(value))
	 }

	return data, nil
}

//create a bulk indexer 
func createBulkIndexer(es *elasticsearch.Client, indexName string) (esutil.BulkIndexer, error){
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         indexName,    
		Client:        es,       
		NumWorkers:    numWorkers,  
		FlushBytes:    int(flushBytes), 
		FlushInterval: 30 * time.Second,
	})

	if err != nil {
		standardLogger.ErrorMessage(err,"Error creating the indexer")
		return nil, err
	}

	return bi, nil
}

//Checking if index exists or not, if doesn't exist create one.
func indexExists(es *elasticsearch.Client, indexName string) {
	resp, err := es.Indices.Exists([]string{indexName})
	if err != nil {
		standardLogger.ErrorMessage(err,"Error while checking for index exists or not")
	}

	if resp.StatusCode != 200 {
		res, err = es.Indices.Create(indexName)
		if err != nil {
			standardLogger.ErrorMessage(err,"Cannot create index")
		}
		if res.IsError() {
			res_error := errors.New(res.String())
			standardLogger.ErrorMessage(res_error,"Cannot create index")
		}

		res.Body.Close()
	} else {
		standardLogger.InfoMessage("Index with this name already exist","")	
	}
}

// inserting data in elastic search
func putQueryResultInSearch(es *elasticsearch.Client, data []string, indexName string){
	// creating bulk indexer
	bi, err := createBulkIndexer(es, indexName)

	if err != nil {
		standardLogger.ErrorMessage(err,"Error while creating bulk indexer")
		return
	}

	//Recording time for inserting data in Elastic search
	start := time.Now().UTC()

	for i, body := range data {
		err = bi.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				// Action field configures the operation to perform (index, create, delete, update)
				Action: "index",

				// Document ID 
				DocumentID: strconv.Itoa(i + 1),

				// Body is the payload
				Body: strings.NewReader(body),

				// OnSuccess is called for each successful operation
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
					atomic.AddUint64(&countSuccessful, 1)
				},

				// OnFailure is called for each failed operation
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						standardLogger.ErrorMessage(err, "")
					} else {
						res_error_type := errors.New(res.Error.Type)
						standardLogger.ErrorMessage(res_error_type, res.Error.Reason)
					}
				},
			},
		)

		if err != nil {
			standardLogger.ErrorMessage(err, "Unexpected result")
		}
	}

	if err := bi.Close(context.Background()); err != nil {
		standardLogger.ErrorMessage(err, "Unexpected error")
	}

	// Report the results: number of indexed docs, number of errors, duration, indexing rate
	biStats := bi.Stats()

	dur := time.Since(start)

	if biStats.NumFailed > 0 {
		standardLogger.BiStatsErrorMessage(humanize.Comma(int64(biStats.NumFlushed)),
			humanize.Comma(int64(biStats.NumFailed)),
			dur.Truncate(time.Millisecond),
			humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed))),)
	} else {
		standardLogger.BiStatsSuccessMessage(humanize.Comma(int64(biStats.NumFlushed)),
			dur.Truncate(time.Millisecond),
			humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed))),)
	}
}

//DB Query Executer
func (server *Server) postQueryResult(w http.ResponseWriter, r *http.Request) {
	
	r.ParseForm()
	dbDriver := r.Form.Get("dbDriver")
	dbName := r.Form.Get("dbName")
	query := r.Form.Get("query")
	indexName := r.Form.Get("indexName")

	standardLogger.InfoMessage("BulkIndexer workers :", strconv.Itoa(numWorkers))
	standardLogger.InfoMessage("BulkIndexer flush :", humanize.Bytes(uint64(flushBytes)))

	//loading env variables
	envVarLoad()
	
	// Initialize ES client
	es, err:=getESClient()

	if err != nil {
		standardLogger.FatalErrorMessage(err, "Error intializing Elastic Search client")
		return
	}

	// checking if index exist or not, if not create one
	indexExists(es, indexName)

	// extracting the query result from database
	data, err:=extractQueryResult(dbDriver, dbName, query)

	if data == nil && err != nil {
		responses.ERROR(w, http.StatusNotImplemented, err)
		standardLogger.ErrorMessage(err, "")
		return
	} else if err != nil {
		StatusCode, errs := strconv.Atoi(data[0])
		if errs != nil {
			standardLogger.ErrorMessage(errs, "Error while converting to int from string")
			return
		}
		standardLogger.ErrorMessage(err, "")
		responses.ERROR(w, StatusCode, err)
		return
	}
	
	//putting data in elastic search
	putQueryResultInSearch(es,data,indexName)

	responses.JSON(w, http.StatusCreated, "{result : posted data successfully on elastic search}")
	
}
