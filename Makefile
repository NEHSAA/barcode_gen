all: build

encrypt_secrets:
	gcloud kms encrypt --plaintext-file=secrets.yaml --location=global --keyring=appengine --key=default --ciphertext-file=secrets.yaml.enc

decrypt_secrets:
	gcloud kms decrypt --plaintext-file=secrets.yaml --location=global --keyring=appengine --key=default --ciphertext-file=secrets.yaml.enc

build:
	go build -o barcode_gen

run:
	go run .

deploy:
	gcloud app deploy

docker:
	docker build -t asia.gcr.io/nehsaa-infra/barcode_gen .

docker_deploy:
	docker push asia.gcr.io/nehsaa-infra/barcode_gen
