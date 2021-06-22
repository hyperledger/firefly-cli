FROM trufflesuite/ganache-cli:latest

ADD healthcheck.sh healthcheck.sh

RUN apk add curl

ENV DOCKER true

EXPOSE 8545

ENTRYPOINT ["node", "/app/ganache-core.docker.cli.js"]