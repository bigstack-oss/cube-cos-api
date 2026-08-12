FROM fedora:35

RUN dnf install -y ca-certificates wget
ENV GOLANG_VERSION=1.24.2
ENV GOTOOLCHAIN=local
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
RUN rm -rf /usr/local/go
RUN wget -q --no-check-certificate https://go.dev/dl/go1.26.5.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.26.5.linux-amd64.tar.gz
RUN rm go1.26.5.linux-amd64.tar.gz

RUN dnf install -y go-task rpmdevtools gh git openssh-clients
RUN dnf install -y make automake gcc gcc-c++ kernel-devel
RUN wget https://github.com/mikefarah/yq/releases/download/v4.44.3/yq_linux_amd64 -O /usr/local/bin/yq && chmod +x /usr/local/bin/yq

ENV USER=jenkins UID=1000 GID=1000
RUN groupadd -g ${GID} ${USER}
RUN useradd -u ${UID} -g ${USER} -d /home/${USER} -s /bin/bash -m ${USER}
RUN mkdir -p /home/${USER}/workspace
RUN chown -R ${USER}:${USER} /home/${USER}/workspace
USER ${USER}

# rpm:prepareTarball runs `git submodule update --init`, and api/cube-cos-openapi is
# registered with an SSH URL. Multibranch tag jobs read their Jenkinsfile from the tag
# commit, so every tag cut before the sshagent fix still reaches github.com with no SSH
# identity and dies on "Permission denied (publickey)". cube-cos-openapi is public, so
# clone it over HTTPS. Scoped to that one repo, which leaves origin on SSH.
RUN git config --global \
    url."https://github.com/bigstack-oss/cube-cos-openapi.git".insteadOf \
    "git@github.com:bigstack-oss/cube-cos-openapi.git"

ENV GOPATH=/home/${USER}/go
ENV PATH=${PATH}:/usr/local/go/bin:${GOPATH}/bin

CMD ["bash"]
