# broker

implemented a simple broker in golang using *grpc*.
reaching more than 30K publishing.

##Proxy
in this project for having connection between client and server i used envoy as proxy.

##DB
in this project i have worked with three databases and used three modules for each
  1. postgres
  2. scylla
  3. cassandra
  
##Monitoring
my scrapper is prometheus and data visualization is grafana. their configs are also in this repo.

##Docker
this project is also dockerized for those who are interseted in. see docker-compose file

##Kubernetes
this project also run on minikube and the point is load balancing and scaling-up was tested on this project and it worked properly(using headless servies)  
