FROM fedora:35

RUN dnf install -y ca-certificates wget
ENV GOLANG_VERSION=1.24.0
ENV GOTOOLCHAIN=local
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
RUN rm -rf /usr/local/go
RUN wget -q --no-check-certificate https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
RUN rm go1.24.0.linux-amd64.tar.gz

RUN dnf install -y go-task rpmdevtools gh

ENV USER=jenkins UID=1000 GID=1000
RUN groupadd -g ${GID} ${USER}
RUN useradd -u ${UID} -g ${USER} -d /home/${USER} -s /bin/bash -m ${USER}
RUN mkdir -p /home/${USER}/workspace
RUN chown -R ${USER}:${USER} /home/${USER}/workspace
USER ${USER}

ENV GOPATH=/home/${USER}/go
ENV PATH=${PATH}:/usr/local/go/bin:${GOPATH}/bin

CMD ["bash"]
