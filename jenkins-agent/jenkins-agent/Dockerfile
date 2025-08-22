FROM jenkins/inbound-agent:latest

USER root

RUN apt-get update && \
    apt-get install -y ansible && \
    apt-get install -y terraform && \
    apt-get clean

USER jenkins