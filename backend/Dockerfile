FROM golang:latest

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o /kubecloud

EXPOSE 8080
CMD [ "/kubecloud" ]