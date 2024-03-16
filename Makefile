devapi:
	@echo "Launching dev API"
	cd server && nodemon --exec go run cmd/server/main.go --signal SIGTERM

devcollector:
	@echo "Launching dev collector"
	cd server && nodemon --exec go run cmd/collector/main.go --signal SIGTERM

tunnel:
	@echo "Launching ngrok tunnel"
	ngrok http https://localhost:8000 --region=eu --hostname=ngrok.rimdian.com

dockerapi:
	@echo "Build server"
	# docker build . --file Dockerfile_api --progress=plain --tag rimdianapi --no-cache
	docker build . --file Dockerfile_api --progress=plain --tag rimdianapi

dockercollector:
	@echo "Build collector"
	# docker build . --file Dockerfile_collector --progress=plain --no-cache
	docker build . --file Dockerfile_collector --progress=plain

cleandocker:
	docker system prune --all --force
	
inspectimage:
	docker exec -it IMAGE_ID /bin/sh