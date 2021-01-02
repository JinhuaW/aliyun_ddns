package main

import (
	"net"
	"flag"
	"strings"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"

)

type RecordInfo struct {
	Name	string	`json:"name"`
	A_R	string	`json:"A"`
	AAAA_R	string	`json:"AAAA"`
	CNAME_R	string	`json:"CNAMEE"`
}

type ConfigInfo struct {
	Domain	string		`json:"domain"`
	KeyId	string		`json:"accessKeyId"`
	Secret	string		`json:"accessSecret"`
	Records []RecordInfo	`json:"record"`
}
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

func main() {
	var ipv6_addr string
	//var ipv4_addr string
	var config ConfigInfo
	flag.Parse()
	data, err := ioutil.ReadFile(*config_file)
	if err != nil {
		fmt.Println("Read config file ", *config_file," failed: ", err)
		return
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return
	}

	//get global IPv6 address
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println("Get IP address failed: ", err)
		return
	}
	for _, address := range addrs {
		if strings.HasPrefix(address.String(),"2408:") {
			ipv6_addr = strings.Split(address.String(), "/")[0]
			break
		}
	}
	//update DNS info
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", config.KeyId, config.Secret)
	if err != nil {
		fmt.Println(err)
		return
	}
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.Scheme = "https"
	request.DomainName = config.Domain

	response, err := client.DescribeDomainRecords(request)
	if err != nil {
		fmt.Println(err)
		return
	}

	var found = false
	for _, rr := range config.Records {
		found = false;
		for _, v := range response.DomainRecords.Record {
			if !found && rr.Name == v.RR {
				if v.Type == "AAAA" {
					if v.Value != ipv6_addr {
						updateRecord(client, v.RecordId, config.Domain, "AAAA", rr.Name, ipv6_addr)
					}
					found = true
				}
			}
		}
		if !found {
			addRecord(client, config.Domain, "AAAA", rr.Name, ipv6_addr)
		}
	}
}
