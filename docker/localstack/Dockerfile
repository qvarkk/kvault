FROM localstack/localstack:latest

ENV DEBIAN_FRONTEND=noninteractive
ENV PIP_ROOT_USER_ACTION=ignore

USER root

RUN apt-get update \
 && apt-get install -y wget unzip ca-certificates \
 && wget -q https://releases.hashicorp.com/terraform/1.9.8/terraform_1.9.8_linux_amd64.zip \
 && unzip terraform_1.9.8_linux_amd64.zip -d /usr/local/bin \
 && rm terraform_1.9.8_linux_amd64.zip \
 && apt-get clean

RUN pip install terraform-local

USER localstack