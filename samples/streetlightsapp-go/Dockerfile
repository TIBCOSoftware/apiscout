FROM alpine:latest
RUN apk update && apk add ca-certificates
ENV HTTPPORT=8090 
ADD streetlightsapp-go .
ADD streetlightsapi.yaml .
EXPOSE 8090
CMD ./streetlightsapp-go