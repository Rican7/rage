# Service definitions
SERVICE_NAME ?= rage
LOCAL_SERVICE_IMAGE ?= rican7/rage
REMOTE_SERVICE_IMAGE ?= us-docker.pkg.dev/rage-398305/docker-images/rage

# Run configs
CONTAINER_PORT ?= 80
HOST_PORT ?= 8081
ENV_FILE ?= .env


all:
	# No default target, for now

local-build:
	docker build -t '${LOCAL_SERVICE_IMAGE}' .

local-run:
	docker run --rm -it --publish ${HOST_PORT}:${CONTAINER_PORT} --env-file ${ENV_FILE} ${LOCAL_SERVICE_IMAGE}

local: local-build local-run

remote-build:
	gcloud builds submit

remote-deploy:
	gcloud run deploy ${SERVICE_NAME} --image ${REMOTE_SERVICE_IMAGE}

remote: remote-build remote-deploy


.PHONY: all local-build local-run local remote-build remote-deploy remote
