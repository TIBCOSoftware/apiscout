FROM alpine:latest
RUN apk update && apk add ca-certificates
ENV HTTPPORT=8080 \ 
    PAYMENTSERVICE=bla
ADD invoiceservice-go .
ADD swagger.json .
ADD asyncapi.yaml .
EXPOSE 8080
CMD ./invoiceservice-go