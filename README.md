# event-manager
Maira event manager

## To deploy Prometheus webhook as cloud function
> cd cloudfunctions 
> 
> go mod vendor

#### Deploying a new cloud function
> gcloud functions deploy <`cloud function name`> --runtime go113 --trigger-http --allow-unauthenticated

#### Updating existing cloud function
> gcloud functions deploy <`cloud function name`> --source=.
