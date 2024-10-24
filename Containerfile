FROM golang:1.22.8-bookworm
 
WORKDIR /app
 
COPY . .
 
RUN go mod download
 
RUN go build

EXPOSE 9143
 
CMD [ “./openstack-usage-exporter ]
