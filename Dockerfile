FROM golang:1.23.0-alpine
WORKDIR /app/
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
ADD /internal/repo/pg/migrations migrations
EXPOSE 8080
RUN cd cmd/app/ && go build -o backend-service
CMD ["cmd/app/backend-service"]