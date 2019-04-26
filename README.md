# Simple web service to get prices for AWS EC2 instances

## Build

```bash
$ go get -m
$ go build ec2price.go
```

## Build as docker image

```bash
$ docker build -t ec2price .
```

## Usage

### Set AWS credentials

```bash
$ aws configure
```

or

```bash
$ export AWS_ACCESS_KEY_ID="YOURACCESSCODEHERE" \
$ export AWS_SECRET_ACCESS_KEY="YourSecretAccessKeyHere0123456789" \
```

>NOTE: Ensure your AWS account has necessary IAM Policy to PriceList resources

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "ec2price",
            "Effect": "Allow",
            "Action": [
                "pricing:DescribeServices",
                "pricing:GetAttributeValues",
                "pricing:GetProducts"
            ],
            "Resource": "*"
        }
    ]
}
```

or simple `AWSPriceListServiceFullAccess` internal AWS managed policy

### Start as binary

```bash
$ ./ec2price 
2019/04/26 16:17:43 Getting prices from AWS
2019/04/26 16:18:15 got prices for 1948 ec2 instance types in 18 regions
2019/04/26 16:18:15 Serving http requests

```

### Start as docker container

```bash
$ docker run \
    -it \
    --rm \
    -e AWS_ACCESS_KEY_ID="YOURACCESSCODEHERE" \
    -e AWS_SECRET_ACCESS_KEY="YourSecretAccessKeyHere0123456789" \
    -p 8000:8000 \
    ec2price
2019/04/26 09:24:52 Getting prices from AWS
2019/04/26 09:24:57 got prices for 1948 ec2 instance types in 18 regions
2019/04/26 09:24:57 Serving http requests

```

## Request prices

### Query price for `m5.large` in `us-east-1` AWS region

```bash
$ curl localhost:8000/us-east-1/m5.large
0.0960000000
```

### Query prices for `us-west-2` AWS Region

```bash
$ curl localhost:8000/us-west-2
{
    "a1.2xlarge": "0.2040000000",
    "a1.4xlarge": "0.4080000000",
    "a1.large": "0.0510000000",
    "a1.medium": "0.0255000000",
    "a1.xlarge": "0.1020000000",
    "c3.2xlarge": "0.4200000000",
    "c3.4xlarge": "0.8400000000",
    "c3.8xlarge": "1.6800000000",

<<< output truncated >>>
```

### Query prices for **all** instances in **all** AWS Regions

```bash
$ curl localhost:8000/all
{
    "ap-northeast-1": {
        "c3.2xlarge": "0.5110000000",
        "c3.4xlarge": "1.0210000000",
        "c3.8xlarge": "2.0430000000",
        "c3.large": "0.1280000000",
        "c3.xlarge": "0.2550000000",
        "c4.2xlarge": "0.5040000000",
        "c4.4xlarge": "1.0080000000",
        "c4.8xlarge": "2.0160000000",
        "c4.large": "0.1260000000",

<<< output skipped >>>

        "z1d.6xlarge": "2.7240000000",
        "z1d.large": "0.2270000000",
        "z1d.metal": "5.4480000000",
        "z1d.xlarge": "0.4540000000"
    },
    "ap-northeast-2": {
        "c4.2xlarge": "0.4540000000",
        "c4.4xlarge": "0.9070000000",
        "c4.8xlarge": "1.8150000000",
        "c4.large": "0.1140000000",
        "c4.xlarge": "0.2270000000",
        "c5.18xlarge": "3.4560000000",

<<< output truncated >>>

```

