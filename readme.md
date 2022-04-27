# dns-gcp-2-yc

Script for migrating DNS zone from Google Cloud Platform to Yandex Cloud.

## Warning

Take a notice that for yandex DNS record names `@.domain.com` and `domain.com` are the same. Merge them to single dns record in google before running scripts.

## How to use

```bash
# export all zones from GCP to dns folder
gcloud dns managed-zones list | awk '{print $1}' | grep -v NAME | xargs -n 1 -I% gcloud dns record-sets export ./dns/% --zone=%

# generate terraform files
go run main.go --dns-dir=./dns --tf-dir=./tf --skip-types=ns,soa

# copy template file and edit if needed
cp main.tf ./tf
cd ./tf


terraform init
terraform plan
terraform apply

```
