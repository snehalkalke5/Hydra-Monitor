package HydraMonitor

import (
	"strconv"
	"fmt"
	"github.com/antigloss/go/logger"
	"net/http"
	"time"
	"github.com/gocarina/gocsv"
	"os"
	"bytes"
	"html/template"
	"log"
)

var count	int
var APIListSlice []Apilist
var LastUpdated	string
type Apilist struct {
	URI    string `csv:"URI"`
	Method string `csv:"METHOD"`
	Body string   `csv:"BODY"`
	Status string `json:"Status,omitempty"`
	ID	int
	Description	string
}

type String_Slice []string

type HTTP_Respond struct{
	APISlice []Apilist
	Description string
}

type DESC_Respond struct{
	Description string
}

func callGet(uri string, index int) {
	start := time.Now()
	var netClient = &http.Client{
 		 Timeout: time.Second * 5,
	}
	res, err := netClient.Get("http://172.17.0.2:8080" + uri)
	if err != nil {
		logger.Error("Error ocurred in GET",err.Error())
		APIListSlice[index].Status = "No Response Found"
	} else {
		if res == nil {
			fmt.Println("Error nahi aaya aur response bhi nahi ",err.Error())

		} else {
			APIListSlice[index].Status = res.Status
			APIListSlice[index].ID = index
			APIListSlice[index].Description = "Dummy"
			end := time.Since(start).String()
			string_time := "This call took like "+end+" seconds"
			logger.Info(string_time)
			res.Body.Close()
		}
	}
}

func callPost(uri string, index int,byteSlice []byte) {
	 b := bytes.NewBuffer(byteSlice)

        client := &http.Client{Timeout: time.Second * 5}

        req, err := http.NewRequest("POST", "http://172.17.0.2:8080"+uri, b)

        if err != nil {
                logger.Error("\n\n Request to Create Request Failed \n\n")
                logger.Error(err.Error())
        }

        req.Close = true
        req.Header.Set("Content-Type", "application/json")
        res, err := client.Do(req)
        if err != nil {
                logger.Error(err.Error())
        }
	if res != nil{
		APIListSlice[index].Status = res.Status
		APIListSlice[index].ID = index
		APIListSlice[index].Description = "Dummy"
		res.Body.Close()
	} else {
		APIListSlice[index].Status = "No Response Found"
	}
	//res.Body.Close()
}


func ShowHTML(w http.ResponseWriter, r *http.Request){	
	t, _ := template.ParseFiles("htmls/index.html") 
	var httpresponse HTTP_Respond
	httpresponse.APISlice = APIListSlice
	httpresponse.Description = getTimeString()
	t.Execute(w, httpresponse)

}

func getTimeString() string{
	return LastUpdated
}


func setTimeString(){
	LastUpdated = time.Now().String()
}

func getDescription(r *http.Request) string{

	mapHTTP := r.URL.Query()
	var Index_Return int64
	Index_Return = -1
	for key, value := range mapHTTP {
		if key == "API" {
			for _, valueStrg := range value {
				Index_Return,_ = strconv.ParseInt(valueStrg, 10, 0)
			}
		}
	}
	returnString := APIListSlice[Index_Return].Description

	return returnString

}

func ShowDescription(w http.ResponseWriter, r *http.Request){
	t, _ := template.ParseFiles("htmls/description.html")
	var descresponse DESC_Respond
	descresponse.Description = getDescription(r)
	t.Execute(w, descresponse)
}	

func StatusServer(){
    mux := http.NewServeMux()
    hf_register := http.HandlerFunc(ShowHTML)
    mux.Handle("/checkStatus", hf_register)
    hf_description := http.HandlerFunc(ShowDescription)
    mux.Handle("/description", hf_description)
    fmt.Println("Server is Now Up")
	go updateData()
    log.Fatal(http.ListenAndServe(":8011", mux))
}


func main(){
	LastUpdated = "Never"
	count = 0
	fmt.Println("Wait for Initial Loading of Data ...  ")
	updateData()
	StatusServer()

}


func updateData() {
	if count != 0 {
		time.Sleep(1 * time.Minute)
	}


	count = count + 1
	file, err := os.Open("./API.csv")
	if err != nil {
		fmt.Println(err)
		return
	}

	 

	if err := gocsv.UnmarshalFile(file, &APIListSlice); err != nil {
		fmt.Println(err)
	}



	defer file.Close()
	

	for index, value := range APIListSlice {
                var temp_apis Apilist
                temp_apis.URI = value.URI
                temp_apis.Method = value.Method
		temp_apis.Body = value.Body	

		bytes_body := []byte(temp_apis.Body)
                if temp_apis.Method == "GET" {
                        callGet(temp_apis.URI, index)
                }  else {
                        callPost(temp_apis.URI, index,bytes_body)
                }
        }
	setTimeString()

}
