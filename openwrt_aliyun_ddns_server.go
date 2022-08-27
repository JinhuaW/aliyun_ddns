package main

import (
	"flag"
	"strings"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"

)

type ConfigInfo struct {
	KeyId	string		`json:"accessKeyId"`
	Secret	string		`json:"accessSecret"`
	Port 	string		`json:"port"`
}

var client *alidns.Client
var config_file = flag.String("c", "aliyun_ddns.json", "aliyun ddns configuration")

func getRecord(client *alidns.Client, domain, rType, rr string) (records []alidns.Record, err error) {
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.Scheme = "https"
	request.DomainName = domain
	request.Type = rType
	request.RRKeyWord = rr

	response, err := client.DescribeDomainRecords(request)
	if err != nil {
		return nil, err
	}
	return response.DomainRecords.Record, nil
}

func addRecord(client *alidns.Client, domain, rType, rr, value string) error {
	request := alidns.CreateAddDomainRecordRequest()
	request.Scheme = "https"
	request.DomainName = domain
	request.Type = rType
	request.RR = rr
	request.Value = value

	_, err := client.AddDomainRecord(request)
	if err != nil {
		return err
	}
	return nil
}

func updateRecord(client *alidns.Client, recordId, domain, rType, rr, value string) error {
	request := alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"
	request.RecordId = recordId
	request.Type = rType
	request.RR = rr
	request.Value = value

	_, err := client.UpdateDomainRecord(request)
	if err != nil {
		return err
	}
	return nil
}

func delRecord(client *alidns.Client, recordId string) error {
	request := alidns.CreateDeleteDomainRecordRequest()
	request.Scheme = "https"
	request.RecordId = recordId
	_, err := client.DeleteDomainRecord(request)
	if err != nil {
		return err
	}
	return nil
}


func dnsHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
		case "GET":
			domain_list := strings.Split(req.FormValue("domain"), ".")
			name := domain_list[0]
			domain := strings.Join(domain_list[1:], ".")
			rType :=  req.FormValue("type")
			ipAddr :=  req.FormValue("ip")

			request := alidns.CreateDescribeDomainRecordsRequest()
			request.Scheme = "https"
			request.DomainName = domain

			response, err := client.DescribeDomainRecords(request)
			if err != nil {
				fmt.Println(err)
				return
			}
			var found = false
			for _, v := range response.DomainRecords.Record {
				if name == v.RR {
					if v.Value != ipAddr {
						updateRecord(client, v.RecordId, domain, rType, name, ipAddr)
					}
					found = true
					break
				}
			}
			if !found {
				addRecord(client, domain, rType, name, ipAddr)
			}
	}
}
func main() {
	var config ConfigInfo
	flag.Parse()
	data, err := ioutil.ReadFile(*config_file)
	if err != nil {
		fmt.Println("Read config file ", *config_file," failed: ", err)
		return
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Println(err)
		return
	}

	//update DNS info
	client, err = alidns.NewClientWithAccessKey("cn-hangzhou", config.KeyId, config.Secret)
	if err != nil {
		fmt.Println(err)
		return
	}
	http.HandleFunc("/", dnsHandler)
	fmt.Println(">> Listen on 127.0.0.1 " + config.Port)
	err = http.ListenAndServe("127.0.0.1:" + config.Port, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

}
