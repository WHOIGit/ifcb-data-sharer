# For more information, please refer to https://aka.ms/vscode-docker-python
FROM python:3.13-slim-bullseye

# Keeps Python from generating .pyc files in the container
ENV PYTHONDONTWRITEBYTECODE=1

# Turns off buffering for easier container logging
ENV PYTHONUNBUFFERED=1

# install Docker in container to execute commands on ifcbdb containers
RUN apt-get update
RUN apt-get -y install apt-transport-https ca-certificates curl gnupg2 software-properties-common
RUN curl -fsSL https://get.docker.com/ | sh

# Install pip requirements
COPY requirements.txt .
RUN python -m pip install -r requirements.txt

WORKDIR /app
COPY . /app

CMD ["python", "app.py"]
