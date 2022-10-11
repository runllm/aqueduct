FROM python:3.8

MAINTAINER Aqueduct <hello@aqueducthq.com> version: 0.0.1

ARG version

USER root

ENV PYTHONUNBUFFERED 1

RUN pip install aqueduct-ml==${version}

CMD aqueduct start --expose

