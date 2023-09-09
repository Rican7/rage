# Service definitions
SERVICE_NAME ?= rage
LOCAL_SERVICE_IMAGE ?= rican7/rage
REMOTE_SERVICE_IMAGE ?= us-docker.pkg.dev/rage-398305/docker-images/rage

# Run configs
CONTAINER_PORT ?= 80
HOST_PORT ?= 8080
ENV_FILE ?= .env


default: ${ENV_FILE}

${ENV_FILE}:
	cp .env.example ${ENV_FILE}

local-app-build:
	docker build -t '${LOCAL_SERVICE_IMAGE}' .

local-app-run: ${ENV_FILE}
	docker run --rm --publish ${HOST_PORT}:${CONTAINER_PORT} --env-file ${ENV_FILE} ${LOCAL_SERVICE_IMAGE}

local-app: local-app-build local-app-run

local:
	@docker compose down \
		|| (echo 'You might be using an old version of Docker Compose. Upgrade to v2.'; exit 1) \
		&& docker compose up

remote-build:
	gcloud builds submit

remote-deploy:
	gcloud run deploy ${SERVICE_NAME} --image ${REMOTE_SERVICE_IMAGE}

remote: remote-build remote-deploy


.PHONY: default local-app-build local-app-run local-app local remote-build remote-deploy remote
