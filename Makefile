container_name=syslog-kafka
binary_name=go-syslog-kafka
docker_registry_account=payneio
docker_tag=latest

build/binary: *.go
	GOOS=linux GOARCH=amd64 go build -o build/$(binary_name)

stage/binary: build/$(binary_name)
	mkdir -p stage
	cp build/$(binary_name) stage/$(binary_name)

build/container: stage/$(binary_name) Dockerfile
	docker build --no-cache -t $(container_name) .
	touch build/container

release:
	docker tag $(container_name) $(docker_registry_account)/$(container_name):$(docker_tag)
	docker push $(docker_registry_account)/$(container_name):$(docker_tag)

.PHONY: clean
clean:
	rm -rf {build,stage}
