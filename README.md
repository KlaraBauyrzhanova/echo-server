# echo-server
This is a simple echo server

## Runnig
```go run main.go```

## Usage
You have 2 endpoints:
- /echo - add new client to sse and once a second sends him the text received from GET;
- /say - takes a word and replaces the word sent from / echo with it for this client.

### Example
- GET /echo?w=hello - returns the word *hello* once a second;
- GET /say?w=world - returns the word *world* once a second.
