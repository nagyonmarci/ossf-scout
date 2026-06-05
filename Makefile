.PHONY: build dev docker clean

build:
	cd frontend && pnpm install --frozen-lockfile && pnpm run build
	go build -o ossf-scout .

# Faster rebuild: skips install if node_modules already exists
dev:
	cd frontend && [ -d node_modules ] || pnpm install --frozen-lockfile
	cd frontend && pnpm run build
	go build -o ossf-scout .

docker:
	DOCKER_BUILDKIT=1 docker compose up --build

clean:
	rm -rf frontend/dist ossf-scout data/
