.PHONY: build dev docker clean

build:
	cd frontend && npm ci && npm run build
	go build -o ossf-scout .

# Faster rebuild: skips npm ci if node_modules already exists
dev:
	cd frontend && [ -d node_modules ] || npm ci
	cd frontend && npm run build
	go build -o ossf-scout .

docker:
	DOCKER_BUILDKIT=1 docker compose up --build

clean:
	rm -rf frontend/dist ossf-scout data/
