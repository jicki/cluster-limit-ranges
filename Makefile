ROOT:=$(shell pwd -P)
GIT_COMMIT:=$(shell git --work-tree ${ROOT}  rev-parse 'HEAD^{commit}')
_GIT_VERSION:=$(shell git --work-tree ${ROOT} describe --tags --abbrev=14 "${GIT_COMMIT}^{commit}" 2>/dev/null)
TAG=$(shell echo "${_GIT_VERSION}" |  awk -F"-" '{print $$1}')

build:
	go build -o cluster-limit-controller main.go

build-image:
	docker build -t="jicki/cluster-limit-controller:$(TAG)" -f Dockerfile .
	docker push jicki/cluster-limit-controller:$(TAG)
