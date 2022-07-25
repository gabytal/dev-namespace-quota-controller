
FROM golang:1.18-alpine AS build

# create a working directory inside the image
WORKDIR /app

# copy Go modules and dependencies to image
COPY go.mod go.sum ./

# download Go modules and dependencies
RUN go mod download

# copy directory files i.e all files ending with .go
COPY *.go ./

# compile application
RUN go build -o /quotacontroller


##
## STEP 2 - DEPLOY
##
FROM alpine

WORKDIR /

COPY configFile /configFile
COPY --from=build /quotacontroller /quotacontroller

ENTRYPOINT ["/quotacontroller"]