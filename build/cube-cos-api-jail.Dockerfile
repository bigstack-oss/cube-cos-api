FROM golang:1.24.2-bookworm

ENV USER=jenkins UID=1000 GID=1000

RUN curl -1sLf 'https://dl.cloudsmith.io/public/task/task/setup.deb.sh' | bash
RUN apt-get update && apt-get install -y task gh build-essential
RUN wget https://github.com/mikefarah/yq/releases/download/v4.44.3/yq_linux_amd64 -O /usr/local/bin/yq && chmod +x /usr/local/bin/yq

ENV USER=jenkins UID=1000 GID=1000
RUN groupadd -g ${GID} ${USER}
RUN useradd -u ${UID} -g ${USER} -d /home/${USER} -s /bin/bash -m ${USER}
RUN mkdir -p /home/${USER}/workspace
RUN chown -R ${USER}:${USER} /home/${USER}/workspace

USER ${USER}
CMD ["bash"]
