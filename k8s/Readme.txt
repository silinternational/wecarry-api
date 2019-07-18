
Assumption:  Minikube has been installed.


# Set the environment variables
eval $(minikube docker-env)

# Create the Docker image and push it to the local Docker repo
cd handcarry-api
make buffalo
make clean

# Deploy to Minikube
kubectl apply -k ./

# Confirm app is up
kubectl get pods
kubectl get deployments
kubectl get services -w

# Access the app
minikube service handcarry --url

# That will give you are URL like this:  http://192.168.39.133:30769
# Navigate to that URL in your web browser.

# On 7/18/2019, I get an error response.  :-(
