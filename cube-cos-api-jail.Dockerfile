FROM golang:1.24-alpine3.21
ENV USER=jenkins UID=1000 GID=1000
RUN apk add --no-cache bash git openssh go-task zip
RUN addgroup -g ${GID} ${USER}
RUN adduser -u ${UID} -G ${USER} -h /home/${USER} -s /bin/bash -D ${USER}
RUN mkdir -p /workspace
RUN chown -R ${USER}:${USER} /workspace
RUN ln -sf /bin/bash /usr/bin/bash
USER jenkins
CMD ["bash"]
