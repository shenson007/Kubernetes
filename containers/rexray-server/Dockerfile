FROM centos

RUN yum update -y
RUN yum install numactl libaio xfsprogs -y

ENV RR_VERSION 0.6.0

RUN curl -sSL https://dl.bintray.com/emccode/rexray/install | sh -s -- stable $RR_VERSION

RUN echo -e '#!/bin/sh\n\n/usr/bin/rexray start -f' > /docker-entrypoint.sh && chmod +x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]
