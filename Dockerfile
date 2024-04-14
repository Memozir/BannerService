FROM golang:1.22

WORKDIR /BannerService

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
ENV GO111MODULE=on


COPY . .
RUN go mod download
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Build
RUN cd cmd/banner_service && go build -o banner-service
EXPOSE 8080
# Run
COPY ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]