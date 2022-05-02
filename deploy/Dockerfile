#FROM amd64/ubuntu:latest
FROM scratch

ENV CONFIG config.yaml
ENV APIKEY  Aa1234
#RUN apt update && apt-get install -y ca-certificates curl wget inetutils-ping ncdu htop iftop

COPY ./compass /compass

ENTRYPOINT [ "/compass" ]