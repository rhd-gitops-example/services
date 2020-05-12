FROM ubuntu
RUN apt-get update

# Install curl
RUN apt-get update 
RUN apt-get install -y curl

# Install kubectl
RUN apt-get install -y apt-transport-https gnupg2
RUN curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
RUN echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | tee -a /etc/apt/sources.list.d/kubernetes.list
RUN apt-get update && apt-get install -y kubectl


# Easy way to install yq, but wastes a lot of image size. Only required for 'standalone' case. 
RUN apt-get install -y software-properties-common 
RUN add-apt-repository -y ppa:rmescandon/yq
RUN apt install -y yq

# Install hub command, and add 'git' alias. 'hub' is a strict superset of 'git'. 
RUN apt install -y hub 
RUN ln -s /usr/local/bin/hub /usr/local/bin/git

RUN rm -rf /var/lib/apt/lists/*