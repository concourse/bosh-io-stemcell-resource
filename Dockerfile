FROM concourse/buildroot:curl

ADD assets/ /opt/resource/

RUN mkdir -p /opt/jq
RUN curl -L https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64 -o /opt/jq/jq

RUN chmod +x /opt/jq/jq
RUN chmod +x /opt/resource/*
