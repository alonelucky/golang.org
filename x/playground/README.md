# playground

This subrepository holds the source for the Go playground:
https://play.golang.org/

## Building

```
# build the image
docker build -t playground .
```

## Running

```
docker run --name=play --rm -d -p 8080:8080 playground
# run some Go code
cat /path/to/code.go | go run client.go | curl -s --upload-file - localhost:8080/compile
```

# Deployment

Building the playground Docker container takes more than the default 10 minute time limit of cloud build, so increase its timeout first (note, `app/cloud_build_timeout` is a global configuration value):

```
gcloud config set app/cloud_build_timeout 1200  # 20 mins
```

Alternatively, to avoid Cloud Build and build locally:

```
make docker
docker tag playground:latest gcr.io/golang-org/playground:latest
docker push gcr.io/golang-org/playground:latest
gcloud --project=golang-org --account=you@google.com app deploy app.yaml --image-url=gcr.io/golang-org/playground:latest
```

Then:

```
gcloud --project=golang-org --account=you@google.com app deploy app.yaml
```

# Contributing

To submit changes to this repository, see
https://golang.org/doc/contribute.html.
