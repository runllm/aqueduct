FROM nvidia/cuda:11.8.0-runtime-ubuntu22.04

MAINTAINER Aqueduct <hello@aqueducthq.com> version: 0.0.1

USER root

RUN apt-get -y update \
    && apt-get install -y wget \
    && apt-get install -y software-properties-common
RUN apt-get -y update

# Install miniconda
ENV CONDA_DIR /opt/conda
RUN wget --quiet https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-x86_64.sh -O ~/miniconda.sh && \
     /bin/bash ~/miniconda.sh -b -p /opt/conda

# Put conda in path so we can use conda activate
ENV PATH=$CONDA_DIR/bin:$PATH

COPY ./gpu/py38_env.yml .
RUN conda init bash && conda env create -f py38_env.yml

ENV PYTHONUNBUFFERED 1

# Download Dolly V2 3B
COPY ./llm/download_model.py .
RUN conda run -n py38_env pip install huggingface_hub
RUN conda run -n py38_env python3 download_model.py --repo-id databricks/dolly-v2-3b --local-dir /dolly-v2-3b

# Install Aqueduct LLM wrapper
RUN apt install git -y
RUN echo a
RUN conda run -n py38_env pip install "git+https://github.com/aqueducthq/aqueduct-llm@vicuna_7b"

WORKDIR /

COPY ./gpu/start-function-executor-gpu.sh /

CMD ["bash","/start-function-executor-gpu.sh", "py38_env"]

