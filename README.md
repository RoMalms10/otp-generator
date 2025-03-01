# otp-generator
Microservice the generates 6 digit OTPs, stores it in a redis cache and sends them via SMS or email. The microservice also validates the OTP

# How to use this microservice
To use this microservice, first start up your redis cache like this: `docker run --name redis -d -p 6379:6379 redis`

Then, start the microservice by using `cd` to get to the `main.go` file and run the command: `go run main.go`

Now, use `curl` to make a request to the microservice to generate an OTP:
```
curl -X POST http://localhost:8080/otp/generate \
     -H "Content-Type: application/json" \
     -d '{
           "username": "user@example.com",
           "messageType": "email"
         }'
```
Expected Response:
```
{
  "otp": "123456"
}
```


Next, take the OTP that was generated and make another request to the microservice to validate it:
```
curl -X POST http://localhost:8080/otp/validate \
     -H "Content-Type: application/json" \
     -d '{
           "username": "user@example.com",
           "otp": "123456"
         }'

```
Expected Response:
```
{
  "status": "success"
}

```
