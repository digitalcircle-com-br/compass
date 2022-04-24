FROM amd64/ubuntu:latest

ENV CONFIG config.yaml

RUN apt update && apt-get install -y ca-certificates curl wget inetutils-ping ncdu htop iftop

COPY ./compass /compass

ENTRYPOINT [ "/compass" ]