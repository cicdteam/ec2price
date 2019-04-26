package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/pricing/pricingiface"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	"net/http"
	"os"
	"log"
	"time"
	"strings"
)

const priceRequestPeriod = 3600 //seconds

const usage = `
AWS EC2 OnDemand prices

Usage:

    curl ec2.cicd.team/{region}/{ec2type}
    curl ec2.cicd.team/{region}
    curl ec2.cicd.team/all

Examples:

    - get hourly onemand price for t3.nano instance in eu-west-1 AWS region
    curl ec2.cicd.team/eu-west-1/t3.nano

    - get hourly onemand prices for all instances in eu-west-1 AWS region
    curl ec2.cicd.team/eu-west-1

    - get hourly onemand prices for all instances in all AWS regions
    curl ec2.cicd.team/all

`

const usage_html = `
<!DOCTYPE html>
<html>
<head>
<title>AWS EC2 OnDeman Prices</title>
</head>
<body>

<span style="font-family:courier new,courier,monospace; font-size:20px;">

<h4>AWS EC2 OnDemand prices</h4>
<p>Usage:</p>
<blockquote>
    <p>
<strong>
    curl ec2.cicd.team/<span style="color: #ff0000;">{AWSregion}</span>/<span style="color: #ff0000;">{EC2type}</span><br />
    curl ec2.cicd.team/<span style="color: #ff0000;">{AWSregion}</span><br />
    curl ec2.cicd.team/<span style="color: #ff0000;">all</span><br />
</strong>
</p>
</blockquote>
<p>Examples:</p>
<blockquote>
<p>
    - get hourly onemand price for <span style="color: #ff0000;">t3.nano</span> instance in <span style="color: #ff0000;">eu-west-1</span> AWS region<br /><strong>
    curl <a href="https://ec2.cicd.team/eu-west-1/t3.nano"><span style="color:#000000;">ec2.cicd.team/</span><span style="color: #ff0000;">eu-west-1</span><span style="color:#000000;">/</span><span style="color: #ff0000;">t3.nano</span></strong></a></p>
<p>
    - get hourly onemand prices for all instances in <span style="color: #ff0000;">eu-west-1</span> AWS region<br /><strong>
    curl <a href="https://ec2.cicd.team/eu-west-1"><span style="color:#000000;">ec2.cicd.team/</span><span style="color: #ff0000;">eu-west-1</span></strong></a></p>
<p>
    - get hourly onemand prices for all instances in all AWS regions<br /><strong>
    curl <a href="https://ec2.cicd.team/all"><span style="color:#000000;">ec2.cicd.team/</span><span style="color: #ff0000;">all</span></strong></a></p>
</blockquote>

</span>

<hr />
<p><code>Sources at <a href="https://github/cicdteam/ec2price">github/cicdteam/ec2price</a></code></p>

</body>
</html>
`

var priceData = map[string]map[string]string{}

var regions = map[string]string{
	"Asia Pacific (Tokyo)":      "ap-northeast-1",
	"Asia Pacific (Seoul)":      "ap-northeast-2",
	"Asia Pacific (Mumbai)":     "ap-south-1",
	"Asia Pacific (Singapore)":  "ap-southeast-1",
	"Asia Pacific (Sydney)":     "ap-southeast-2",
	"Canada (Central)":          "ca-central-1",
	"EU (Frankfurt)":            "eu-central-1",
	"EU (Stockholm)":            "eu-north-1",
	"EU (Ireland)":              "eu-west-1",
	"EU (London)":               "eu-west-2",
	"EU (Paris)":                "eu-west-3",
	"South America (Sao Paulo)": "sa-east-1",
	"US East (N. Virginia)":     "us-east-1",
	"US East (Ohio)":            "us-east-2",
	"US West (N. California)":   "us-west-1",
	"US West (Oregon)":          "us-west-2",
	"AWS GovCloud (US-East)":    "us-gov-east-1",
	"AWS GovCloud (US)":         "us-gov-west-1",
}

func getPrices(svc pricingiface.PricingAPI) error {
	var priceInput = &pricing.GetProductsInput{
		Filters: []*pricing.Filter{
			{
				Field: aws.String("ServiceCode"),
				Value: aws.String("AmazonEC2"),
				Type:  aws.String(pricing.FilterTypeTermMatch),
			},
			{
				Field: aws.String("operatingSystem"),
				Value: aws.String("Linux"),
				Type:  aws.String(pricing.FilterTypeTermMatch),
			},
			{
				Field: aws.String("preInstalledSw"),
				Value: aws.String("NA"),
				Type:  aws.String(pricing.FilterTypeTermMatch),
			},
			{
				Field: aws.String("licenseModel"),
				Value: aws.String("No License required"),
				Type:  aws.String(pricing.FilterTypeTermMatch),
			},
			{
				Field: aws.String("capacitystatus"),
				Value: aws.String("UnusedCapacityReservation"),
				Type:  aws.String(pricing.FilterTypeTermMatch),
			},
			{
				Field: aws.String("tenancy"),
				Value: aws.String("Shared"),
				Type:  aws.String(pricing.FilterTypeTermMatch),
			},
		},
		FormatVersion: aws.String("aws_v1"),
		MaxResults:    aws.Int64(100),
		ServiceCode:   aws.String("AmazonEC2"),
	}

	priceDataNew := make(map[string]map[string]string)

	err := svc.GetProductsPages(priceInput,
		func(page *pricing.GetProductsOutput, lastPage bool) bool {
			var reg, instanceType, price string
			for _, item := range page.PriceList {
				reg = regions[item["product"].(map[string]interface{})["attributes"].(map[string]interface{})["location"].(string)]
				if reg == "" {
					continue
				}
				instanceType = item["product"].(map[string]interface{})["attributes"].(map[string]interface{})["instanceType"].(string)
				ondemand_map := item["terms"].(map[string]interface{})["OnDemand"].(map[string]interface{})
				for _, dimen := range ondemand_map {
					for _, pricedim := range dimen.(map[string]interface{})["priceDimensions"].(map[string]interface{}) {
						price = pricedim.(map[string]interface{})["pricePerUnit"].(map[string]interface{})["USD"].(string)
					}
				}

				if priceDataNew[reg] == nil {
					priceDataNew[reg] = map[string]string{}
				}
				priceDataNew[reg][instanceType] = price
			}
			return true
		})

	if err == nil {
		priceData = priceDataNew
	}

	return err
}

func getPricesLoop(svc pricingiface.PricingAPI) {
	for {
		time.Sleep(priceRequestPeriod * time.Second)
		err := getPrices(svc)
		if err != nil {
			log.Println("ERROR:", err.Error())
		}
	}
}

func answerPrice(w http.ResponseWriter, r *http.Request) {

	var region, ec2type string
	region = mux.Vars(r)["region"]
	ec2type = mux.Vars(r)["ec2type"]

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")

	if len(region) > 0 {
		if len(priceData[region]) > 0 {
			if len(ec2type) > 0 {
				w.Header().Set("Content-Type", "text/html; charset=UTF-8")
				if len(priceData[region][ec2type]) > 0 {
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, priceData[region][ec2type])
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			} else {
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(http.StatusOK)
				enc := json.NewEncoder(w)
				enc.SetIndent("", "    ")
				enc.Encode(priceData[region])
			}
		} else {
			w.Header().Set("Content-Type", "text/html; charset=UTF-8")
			w.WriteHeader(http.StatusNotFound)
		}
	} else {
		if len(priceData) > 0 {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusOK)
			enc := json.NewEncoder(w)
			enc.SetIndent("", "    ")
			enc.Encode(priceData)
		} else {
			w.Header().Set("Content-Type", "text/html; charset=UTF-8")
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func usagePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if strings.Index(r.Header.Get("User-Agent"), "curl") > -1 {
		fmt.Fprintln(w, usage)
	} else {
		fmt.Fprintln(w, usage_html)
	}
}

type Route struct {
	Name        string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"allPrices",
		"/all",
		answerPrice,
	},
	Route{
		"regionPrices",
		"/{region}",
		answerPrice,
	},
	Route{
		"instancePrice",
		"/{region}/{ec2type}",
		answerPrice,
	},
	Route{
		"index",
		"/",
		usagePage,
	},
}

func NewRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods("GET", "HEAD").
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}
	return handlers.CombinedLoggingHandler(os.Stdout, handlers.ProxyHeaders(router))
}

func awsInit() *session.Session {
	awsconf := &aws.Config{
		Region: aws.String("us-east-1"), // use us-east-1 as aws pricing api available only in two regions
		CredentialsChainVerboseErrors: aws.Bool(false),
	}
	sess := session.Must(session.NewSession(awsconf))
	return sess
}

func main() {

	log.Println("Getting prices from AWS")

	svc := pricing.New(awsInit())
	err := getPrices(svc)
	if err != nil {
		log.Println("ERROR:", err.Error())
	} else {
		rsum := 0
		isum := 0
		for i := range priceData {
			rsum++
			isum = isum + len(priceData[i])
		}
		log.Println("got prices for", isum, "ec2 instance types in", rsum, "regions")
	}

	go getPricesLoop(svc)

	log.Println("Serving http requests")
	router := NewRouter()

	srv := &http.Server{
		Addr:           ":8000",
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        router,
	}

	log.Fatal(srv.ListenAndServe())
}
