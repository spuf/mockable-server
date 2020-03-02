# Mockable Server

Simple programmable HTTP server helps to mock external services in your tests.

## How it works

There are 2 HTTP servers: first is mock on port 8010, second is control on 8020.

Any request to mock server stores to _Requests_ queue, and sends back data from _Responses_ queue, or HTTP 501.

## Configuration

```shell
$ docker run --rm spuf/mockable-server --help
Usage of mockable-server:
  -control-addr string
        Control server address [CONTROL_ADDR] (default ":8020")
  -mock-addr string
        Mock server address [MOCK_ADDR] (default ":8010")
```

## Usage example

docker-compose.yml:
```yaml
services:
    mockable-server:
        image: spuf/mockable-server:latest
        environment:
            MOCK_ADDR: :8010 
            CONTROL_ADDR: :8020   

    your-service:
        environment:
            TEST_MOCK_SERVER_HOST: mockable-server
            TEST_MOCK_SERVER_PORT: 8010
            TEST_MOCK_SERVER_CONTROL_PORT: 8020
        depends_on:
            - mockable-server
```

## Control API

Has health check endpoint `:8020/healthz`.

Uses JSON-API 1.0 at `:8020/rpc/1`.

### Requests queue

Show queue content:
```json
{
    "method": "Requests.List",
    "params": []    
}
``` 
```json
{
    "id": null,
    "result": [
        {
            "method": "GET",
            "url": "/requested-path?n=v",
            "headers": {
                "Accept": "*/*",
                "Accept-Encoding": "gzip, deflate",
                "Connection": "keep-alive"
            },
            "body": ""
        },
        {
            "method": "POST",
            "url": "/requested-path",
            "headers": {
                "Accept": "*/*",
                "Accept-Encoding": "gzip, deflate",
                "Connection": "keep-alive",
                "Content-Type": "application/x-www-form-urlencoded"
            },
            "body": "n=v"
        }
    ],
    "error": null
}
``` 

Clear queue content:
```json
{
    "method": "Requests.Clear",
    "params": []    
}
```     
```json
{
    "result": true,
    "error": null
}
``` 

Pop request:
```json
{
    "method": "Requests.Pop",
    "params": []    
}
```    
```json
{
    "result": {
        "method": "GET",
        "url": "/requested-path?n=v",
        "headers": {
            "Accept": "*/*",
            "Accept-Encoding": "gzip, deflate",
            "Connection": "keep-alive"
        },
        "body": ""
    },
    "error": null
}
``` 

### Responses queue

Show queue content:
```json
{
    "method": "Responses.List",
    "params": []    
}
```    
```json
{
    "id": null,
    "result": [
        {
            "status": 200,
            "headers": {
                "Content-Type": "text/plain",
                "Extra-Header": "value"
            },
            "body": "OK"
        },
        {
            "status": 200,
            "headers": {
                "Content-Type": "text/plain",
                "Extra-Header": "value"
            },
            "body": "OK"
        }
    ],
    "error": null
}
``` 

Clear queue content:
```json
{
    "method": "Responses.Clear",
    "params": []        
}
``` 
```json
{
    "result": true,
    "error": null
}
``` 

Push response:
```json
{
    "method": "Responses.Push",
    "params": [{
        "status": 200,
        "headers": {
            "Content-Type": "text/plain",
            "Extra-Header": "value"
        },
        "body": "OK"
    }]    
}              
``` 
```json
{
    "result": true,
    "error": null
}
``` 
