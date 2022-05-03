#!/bin/bash

docker run -v $(pwd)/pg-data:/var/lib/postgresql:z --name pg -e POSTGRES_PASSWORD=secretpass -p 5432:5432 -d postgres
