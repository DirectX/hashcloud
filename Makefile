IMAGE=directx/hashcloud
VERSION=$(shell date '+%Y%m%d-%H%M%S')

.DEFAULT: build-ui build push

build-ui:
	@cd frontend && npm install && npm run build && cd ..

build:
	@echo "Building docker image: ${IMAGE}:${VERSION}"
	@docker build -t ${IMAGE}:${VERSION} .
	@docker tag ${IMAGE}:${VERSION} ${IMAGE}:latest
	@echo
	@echo "The image has been built: ${IMAGE}:${VERSION}"
	@echo

push:
	@docker push ${IMAGE}