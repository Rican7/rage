# Service definitions
SERVICE_NAME ?= rage
LOCAL_SERVICE_IMAGE ?= rican7/rage
REMOTE_SERVICE_IMAGE ?= us-docker.pkg.dev/rage-398305/docker-images/rage

# Run configs
CONTAINER_PORT ?= 80
HOST_PORT ?= 8081
ENV_FILE ?= .env


default: ${ENV_FILE}

${ENV_FILE}:
	cp .env.example ${ENV_FILE}

local-build:
	docker build -t '${LOCAL_SERVICE_IMAGE}' .

local-run: ${ENV_FILE}
	docker run --rm --publish ${HOST_PORT}:${CONTAINER_PORT} --env-file ${ENV_FILE} ${LOCAL_SERVICE_IMAGE}

local: local-build local-run

remote-build:
	gcloud builds submit

remote-deploy:
	gcloud run deploy ${SERVICE_NAME} --image ${REMOTE_SERVICE_IMAGE}

remote: remote-build remote-deploy


.PHONY: default local-build local-run local remote-build remote-deploy remote
