FROM concourse/buildroot:curl

ADD assets/ /opt/resource/
RUN chmod +x /opt/resource/*
