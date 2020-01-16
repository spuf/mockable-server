# Mockable Server

Simple programmable HTTP server helps to mock external services in your tests.

## How it works

There are 2 HTTP servers: first is mock on port 8010, second is control on 8020.

Any request to mock server stores to _Requests_ queue, and sends back data from _Responses_ queue, or HTTP 501.

## Control API

Uses JSON-API 1.0 at `:8020/rpc/1`

### Requests queue

Show queue content:
```json
{
    "method": "Requests.List",
    "params": []    
}
``` 

Clear queue content:
```json
{
    "method": "Requests.Clear",
    "params": []    
}
``` 

Pop request:
```json
{
    "method": "Requests.Pop",
    "params": []    
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

Clear queue content:
```json
{
    "method": "Responses.Clear",
    "params": []        
}
``` 

Push response:
```json
{
    "method": "Responses.Push",
    "params": [{
        "code": 200,
        "headers": {
            "Content-Type": "text/plain"
        },
        "body": "OK"
    }]    
}
``` 
