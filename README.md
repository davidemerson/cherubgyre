# cherubgyre
cherubgyre is an anonymous community defense social network

https://cherubgyre.com is under construction, but it's got some links.
https://api.cherubgyre.com has api docs.

The api is deployed on [Digital Ocean](http://64.227.1.200:8080/)

Relaunched in golang!


TO BUILD THIS : `GOOS=linux GOARCH=amd64 go build -o main main.go`



# Getting Started

Clone the docker image
```
docker pull umerfarooq478/cherubgyre
```

After that, run the image
```
docker run -d -p 80:8080 --name cherubgyre-container umerfarooq478/cherubgyre

```

# Contributing

After implementing the changes, update the Dockerfile if needed. And build it
```
docker build --platform linux/amd64 -t umerfarooq478/cherubgyre:latest .
```

After building the image, test it by running

```
docker run -d -p 80:8080 --name cherubgyre-container umerfarooq478/cherubgyre

```

Then call the home api
```
curl http://localhost:80/
```

This should return You've reached cherubgyre

Test All your apis and create a pull request
