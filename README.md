# broker

implemented a simple broker in golang using *grpc*.
reaching more than 30K publishing.

## Proxy
in this project, for having a connection between client and server I used envoy as a proxy and also a rate-limiter.

## DB
In this project, I have worked with three databases and used three modules for each
  1. postgres
  2. Scylla
  3. Cassandra
  
## Monitoring
my scrapper is Prometheus and my data visualization is grafana. their configs are also in this repo.

## Docker
this project is also dockerized for those who are interested. See docker-compose file

## Kubernetes
this project also runs on Minikube and the point is load balancing and scaling-up were tested on this project and it worked properly(using headless services)  
