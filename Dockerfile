FROM golang:1.8

ENV GLIDE_VERSION 0.12.3

# RUN apk add --update curl && \
#     apk add --update git && \
#     apk add --update build-base && \
#     rm -rf /var/cache/apk/*

RUN apt-get update \
    && apt-get install -y \
        git \
    && apt-get clean \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://github.com/Masterminds/glide/releases/download/v$GLIDE_VERSION/glide-v$GLIDE_VERSION-linux-amd64.tar.gz -o glide.tar.gz \
  && tar -xzf glide.tar.gz \
	&& mv linux-amd64/glide /usr/local/bin \
	&& rm -rf linux-amd64 \
	&& rm glide.tar.gz

RUN go get -tags nopq -tags nomysql -tags nomymysql github.com/steinbacher/goose/cmd/goose

RUN mkdir -p /usr/local/go/src/github.com/chrisseto/asset-server
WORKDIR /usr/local/go/src/github.com/chrisseto/asset-server

COPY ./glide.lock /usr/local/go/src/github.com/chrisseto/asset-server
COPY ./glide.yaml /usr/local/go/src/github.com/chrisseto/asset-server

RUN glide install
RUN glide rebuild

COPY ./ /usr/local/go/src/github.com/chrisseto/asset-server

RUN goose up

RUN make build

CMD ["./build/asset-server"]
