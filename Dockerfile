FROM golang:alpine

COPY . /fsrv
RUN cd /fsrv && go build

EXPOSE 9393
CMD ["./fsrv"]
