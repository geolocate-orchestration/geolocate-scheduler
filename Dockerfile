FROM golang:1.16-alpine

WORKDIR /code

COPY . .
RUN go build && mv aida-scheduler /aida-scheduler

WORKDIR /
RUN rm -rf /code

CMD /aida-scheduler
