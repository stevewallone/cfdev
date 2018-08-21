FROM golang

ARG CF_DEPLOY_VER
ENV CF_DEPLOY_VER ${CF_DEPLOY_VER:-3.4}
ENV CF_CLI_VERSION 6.38.0
ENV GOPATH=/go

RUN set -x && \
  curl -fsSL "https://packages.cloudfoundry.org/stable?release=linux64-binary&version=$CF_CLI_VERSION&source=github-rel" \
      | tar -xz -C /usr/bin cf && \
  chmod a+x /usr/bin/cf

RUN set -x && \
  go get -d github.com/cloudfoundry/cf-acceptance-tests && \
  cd /go/src/github.com/cloudfoundry/cf-acceptance-tests && \
  git checkout cf${CF_DEPLOY_VER} && \
  ./bin/update_submodules && \
  git submodule update --init --recursive

COPY ./meow.sh /meow.sh
ENTRYPOINT /meow.sh
