package main

import (
	"log"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	lh "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
)

const (
	TrafficNormal = "NETWORK_NORMAL"
)

type LighthouseClient interface {
	RunningInstances() []string
	ListTrafficPackages() []TrafficPackage
	ShutdownInstance(instanceID string) bool
}

type lighthouseClient struct {
	secretID  string
	secretKey string
	region    string
}

func NewLighthouseClient(id, key, region string) LighthouseClient {
	return &lighthouseClient{
		secretID:  id,
		secretKey: key,
		region:    region,
	}
}

type TrafficPackage struct {
	InstanceID string
	Total      int64
	Used       int64
}

func (pkg *TrafficPackage) UseRate() float64 {
	if pkg.Used <= 0 {
		return 0
	}
	if pkg.Total == 0 {
		return 1
	}

	return float64(pkg.Used) / float64(pkg.Total)
}

func (c *lighthouseClient) RunningInstances() []string {
	limit := int64(100)
	req := lh.NewDescribeInstancesRequest()
	req.Limit = &limit
	res, err := c.apiClient().DescribeInstances(req)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Printf("[error] An API error has returned: %s", err)
		return nil
	}
	if err != nil {
		panic(err)
	}
	var result []string

	for _, instance := range res.Response.InstanceSet {
		if instance.InstanceState != nil && *instance.InstanceState == "RUNNING" {
			result = append(result, *instance.InstanceId)
		}
	}
	// log.Printf("[INFO] There are %d running instances: %v", len(result), result)
	return result
}

func (c *lighthouseClient) ListTrafficPackages() []TrafficPackage {
	limit := int64(100)
	req := lh.NewDescribeInstancesTrafficPackagesRequest()
	req.Limit = &limit
	res, err := c.apiClient().DescribeInstancesTrafficPackages(req)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Printf("[error] An API error has returned: %s", err)
		return nil
	}
	if err != nil {
		panic(err)
	}

	runningInstances := c.RunningInstances()
	if len(runningInstances) < 1 {
		log.Printf("[INFO] you don't have any running instance.\n")
		return nil
	}
	var result []TrafficPackage

	for _, instance := range res.Response.InstanceTrafficPackageSet {
		if instance.InstanceId != nil && sliceContains(runningInstances, *instance.InstanceId) {
			p := TrafficPackage{}
			p.InstanceID = *instance.InstanceId
			for _, pkg := range instance.TrafficPackageSet {
				if packageInUse(pkg) {
					p.Used += *pkg.TrafficUsed
					p.Total += *pkg.TrafficPackageTotal
				}
			}
			result = append(result, p)
		}
	}
	return result
}

func (c *lighthouseClient) ShutdownInstance(instanceID string) bool {
	req := lh.NewStopInstancesRequest()
	req.InstanceIds = common.StringPtrs([]string{instanceID})

	_, err := c.apiClient().StopInstances(req)

	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Printf("[error] An Api error has returned: %s", err)
		return false
	}

	if err != nil {
		log.Printf("[error] Stop lighthouse(%s) server failed. Error: %s", instanceID, err)
		return false
	}

	return true
}

func (c *lighthouseClient) apiClient() *lh.Client {
	credential := common.NewCredential(c.secretID, c.secretKey)
	cpf := profile.NewClientProfile()
	client, _ := lh.NewClient(credential, c.region, cpf)
	return client
}

func packageInUse(pkg *lh.TrafficPackage) bool {
	// now := time.Now()
	if *pkg.Status != TrafficNormal {
		return false
	}

	/*if pkg.StartTime != nil {
		start, _ := time.Parse(TimeLayout, *pkg.StartTime)
		if start.After(now) {
			return false
		}
	}

	if pkg.EndTime != nil {
		end, _ := time.Parse(TimeLayout, *pkg.EndTime)
		if end.Before(time.Now()) {
			return false
		}
	}*/

	return true
}

func sliceContains(haystack []string, needle string) bool {
	for _, str := range haystack {
		if str == needle {
			return true
		}
	}

	return false
}
