# Description

My solution to the fetch backend apprentice coding challenge. A small webserver written in go that fulfills the requirements of the fetch coding challenge. I was an intern at fetch last summer so my solution is inspired by what I learned on while on Starter pack. If there are any issues when running the project feel free to email me at psmccarty@wisc.edu.

## Project Setup

Requires go or docker to run. After cloning the repo you must be in the `fetch-backend-apprentice-challenge` directory.

## Starting the service

### From Local Machine
1. Run setup commands
```
go mod tidy
go mod vendor
```
2. Run the command on the command line
```
go run cmd/application/main.go
```

### Using Docker
1. Build the container
```
docker build . -t pats-receipt-service
```
2. Run the container
```
docker run -it -p 8080:8080 pats-receipt-service
```
---
Now you should have the server listening on port 8080. Use `ctrl-c` to kill it.

## Interacting with the server
The server responds to the two requests specified by the assignment description. You can use Postman or a UNIX shell (worked on macOS and I assume on Linux) using these two sample `curl`s (you can change the json data as much as you like):

1. Path: `/receipts/process` Method: POST. Used for storing receipts.
```
curl --location 'http://localhost:8080/receipts/process' \
--header 'Content-Type: application/json' \
--data '{
  "retailer": "Target",
  "purchaseDate": "2022-01-01",
  "purchaseTime": "13:01",
  "items": [
    {
      "shortDescription": "Mountain Dew 12PK",
      "price": "6.49"
    },{
      "shortDescription": "Emils Cheese Pizza",
      "price": "12.25"
    },{
      "shortDescription": "Knorr Creamy Chicken",
      "price": "1.26"
    },{
      "shortDescription": "Doritos Nacho Cheese",
      "price": "3.35"
    },{
      "shortDescription": "   Klarbrunn 12-PK 12 FL OZ  ",
      "price": "12.00"
    }
  ],
  "total": "35.35"
}'
```

2. Path: `/receipts/{id}/points` Method: GET. Used for getting the points total. Where you replace `{id}` with the id of the receipt that is returned from the previous endpoint.
```
curl --location 'http://localhost:8080/receipts/{id}/points'
```

## Running unit tests

To run the unit tests I have written for `handler.go` use the command:
```
go test ./...
```


