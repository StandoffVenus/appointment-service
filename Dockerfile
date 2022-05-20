FROM golang:1.18.2 AS build
WORKDIR /app/src

RUN apt-get update
RUN apt-get install -y make

COPY go.mod go.sum .
RUN go mod download

COPY . .
RUN make install

# Using this stage allows us to avoid downloading
# and recompiling the project when build args change.
FROM build AS final
ARG DB_NAME=db.sqlite3
ARG PORT=8080

RUN make DB_NAME=${DB_NAME} seed

ENV DB_NAME=${DB_NAME}
ENV PORT=${PORT}
ENTRYPOINT bin/server --file=${DB_NAME} --port=${PORT}
