package main

import (
	"net"
	"strings"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"

)
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
	var ipv4_addr string
	RRs := []string{"@", "www", "git", "cloud", "share", "dl", "iot", "rt", "oa", "mail"}
	domainName := "<your domain>"
	//get global IPv6 address
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println(err)
		return
	}
	for _, address := range addrs {
		if strings.HasPrefix(address.String(),"2408:") {
			ipv6_addr = strings.Split(address.String(), "/")[0]
			break
		}
	}

	//get global IPv4 address
	addr4s, err := net.LookupHost("<oray domain(for ipv4)>")
	if err != nil || addr4s[0] == "0.0.0.0" {
		ipv4_addr = ""
	} else {
		ipv4_addr = addr4s[0]
	}
	//update DNS info
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", "<accessKeyId>", "<accessSecret>")
	if err != nil {
		fmt.Println(err)
		return
	}
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.Scheme = "https"
	request.DomainName = domainName

	response, err := client.DescribeDomainRecords(request)
	if err != nil {
		fmt.Println(err)
		return
	}
	var record4 *alidns.Record
	var record6 *alidns.Record
	for _, rr := range RRs {
		record4 = nil
		record6 = nil
		for _, v := range response.DomainRecords.Record {
			if rr == v.RR && v.Type == "AAAA" {
				if record6 != nil {
					delRecord(client, v.RecordId)
				} else {
					record6 = &v
					if v.Value != ipv6_addr {
						updateRecord(client, v.RecordId, domainName, "AAAA", rr, ipv6_addr)
					}
				}
			}
			if ipv4_addr != "" && rr == v.RR && v.Type == "A" {
				if record4 != nil {
					delRecord(client, v.RecordId)
				} else {
					record4 = &v
					if v.Value != ipv4_addr {
						updateRecord(client, v.RecordId, domainName, "A", rr, ipv4_addr)
					}
				}
			}
		}
		if ipv4_addr != "" && record4 == nil {
			addRecord(client, domainName, "A", rr, ipv4_addr)
		}
		if record6 == nil {
			addRecord(client, domainName, "AAAA", rr, ipv6_addr)
		}
	}
}
