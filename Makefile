.PHONY: build docker clean

build:
	cd frontend && npm ci && npm run build
	go build -o ossf-scout .

docker:
	docker compose up --build

clean:
	rm -rf frontend/dist ossf-scout data/
