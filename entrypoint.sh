#!/bin/sh

export GOOSE_DBSTRING="user=postgres password=postgres port=5432 host=storage dbname=banner_db sslmode=disable"
export  GOOSE_DRIVER=postgres
goose -dir migrations/ up
./cmd/banner_service/banner-service