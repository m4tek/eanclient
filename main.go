package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/tarm/serial"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"os"
	"github.com/joho/godotenv"
)

var apikey string
var hahost string
var eanserver string

type Product struct {
	Id           int     `json:”id”`
	Name         string  `json:”name”`
	EAN          string  `json:”EAN”`
	Category     string  `json:"Category"`
	SimpleName   string  `json:"Simplename"`
	Manufacturer string  `json:"Manufacturer"`
	Amount       float32 `json:"Amount"`
	Unit         string  `jsoon:"Unit"`
	Brand        string  `json:"Brand"`
}

func getEnvVars() {
       err := godotenv.Load("config.env")
       if err != nil {
               log.Fatal("Error loading .env file")
       }
}


func updateListByREST(name string, client *http.Client) {
	url := "https://"+hahost+"/api/services/shopping_list/add_item"

	payload := strings.NewReader("{ \"name\": \"" + name + "\" }")

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Authorization", "Bearer " + apikey)
	req.Header.Add("Content-Type", "application/json")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(string(body))

}

func getProductByREST(ean string, product Product, client *http.Client) string {
	res, err := client.Get("https://"+eanserver+"/product/ean/" + ean)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		log.Print(err)
		return ""
	} else if res.StatusCode != 200 {
		log.Print("Got error: " + string(http.StatusText(res.StatusCode)))
	} else {
		var result map[string]interface{}

		err := json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			log.Print(err)
		} else {
			log.Println(result)
			return result["Simplename"].(string)
		}
	}
	return ""
}

func main() {

	getEnvVars()

	apikey = os.Getenv("HOMEASSISTANT_APIKEY")
	hahost = os.Getenv("HA_HOSTNAME")
	eanserver = os.Getenv("EANSERVER_HOSTNAME")

	serialport := os.Getenv("SERIAL_PORT")

	log.Println(apikey)
	log.Println(hahost)
	log.Println(eanserver)
	log.Println(serialport)

	config := &serial.Config{
		Name: "/dev/ttyACM1",
		Baud: 9600,
		// ReadTimeout: 1,
		// Size: 8,
	}

	//this is server.crt from eanserver (https://github.com/m4tek/eanserver)

	caCert, err := ioutil.ReadFile("server.crt")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	stream, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(stream)
	for {
		text, _ := reader.ReadString('\n')

		text = strings.TrimSpace(text)

		fmt.Printf("Got string %s from reader, fetching data from EAN server\n", string(text))

		var product Product

		simplename:=getProductByREST(text, product, client)

		if simplename!="" {
			updateListByREST(simplename, client)
		} else {
			//TODO some error handling for missing product.
			// maybe reconfigure the scanner to make a good beep when item is found
			// and a bad beep when it is not found, but this needs more serial
			// handling and trasmitting chars back to serial port
		}

	}
}
