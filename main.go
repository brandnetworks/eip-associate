package main

import (
	"flag"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
	"fmt"
	"strings"
	"net/http"
	"io/ioutil"
	"time"
)

// associates an elastic ip with an ec2 instance, picking a free one from a predefined list
func main() {

	flag.Usage = func() {
		fmt.Printf("Usage: eip-associate --eips eips\n")
		flag.PrintDefaults()
	}

	eips := flag.String("eips", "", "Comma separated list of elastic ips")
	maxRetries := flag.Int("retries", 10, "Maximum number of retries")
	pause := flag.Int("pause", 5, "Number of seconds to pause between retries")
	metadata := flag.String("metadata", "http://169.254.169.254/latest/meta-data", "Meta data endpoint")
	flag.Parse()

	if *eips == "" {
		flag.Usage()
		return
	}

	httpClient := http.DefaultClient

	availabilityZoneMetadataEndpoint := *metadata + "/placement/availability-zone"

	log.Println("connecting", availabilityZoneMetadataEndpoint)

	az, err := requestContent(httpClient, availabilityZoneMetadataEndpoint)
	if err != nil {
		log.Fatal(err)
	}

	elasticIps := strings.Split(*eips, ",")

	awsConfig := aws.NewConfig().WithRegion((*az)[:len(*az)-1])
	svc := ec2.New(session.New(), awsConfig)

	log.Println("connected", *az)

	publicIps := make(map[string]struct{})
	for _, ip := range elasticIps {
		publicIps[ip] = struct{}{}
	}

	req := ec2.DescribeAddressesInput{}

	resp, err := svc.DescribeAddresses(&req)

	if err != nil {
		log.Fatal(err)
	}

	instanceIdMetadataEndpoint := *metadata + "/instance-id"

	log.Println("connecting", instanceIdMetadataEndpoint)

	instanceId, err := requestContent(httpClient, instanceIdMetadataEndpoint)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(*instanceId)

	publicIpv4MetadataEndpoint := *metadata + "/public-ipv4"

	publicIpv4, err := requestContent(httpClient, publicIpv4MetadataEndpoint)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(*publicIpv4)

	for _, ip := range elasticIps {
		if *publicIpv4 == ip {
			log.Println("ip already allocated", *publicIpv4, ip)
			return;
		}
	}

	retries := 0;

	for _, address := range resp.Addresses {
		if retries > *maxRetries {
			log.Fatal("Unable to associate public ip")
		}
		_, ok := publicIps[*address.PublicIp]
		if ok {
			if isEipFree(address) {
				log.Println(*address.PublicIp, "free")
				// attempt association
				associateAddressReq := ec2.AssociateAddressInput{AllocationId: address.AllocationId, InstanceId: instanceId}
				_, err := svc.AssociateAddress(&associateAddressReq)
				if err != nil {
					log.Println(err)
				} else {
					log.Println(*address.PublicIp, "associated")
					break;
				}
			} else {
				log.Println(*address.PublicIp, "not_free")
			}
		}
		sleepTime := time.Duration(*pause) * time.Second
		time.Sleep(sleepTime)
	}

}

func isEipFree(address *ec2.Address) (bool) {
	return (address.InstanceId == nil || address.AllocationId == nil)
}

// request the content of a http endpoint as a string
func requestContent(client *http.Client, endpoint string) (*string, error) {
	resp, err := client.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	value, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := string(value)

	return &result, nil
}
