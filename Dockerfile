FROM golang:1.22

WORKDIR /BannerService

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
ENV GO111MODULE=on

COPY . .
RUN go mod download

# Build
RUN cd cmd/banner_service && go build -o banner-service


EXPOSE 8080

# Run
CMD ["./cmd/banner_service/banner-service"]