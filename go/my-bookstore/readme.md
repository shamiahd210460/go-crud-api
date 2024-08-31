We first built a Docker image locally and pushed it to Docker Hub. To test it, we ran a container with the command `docker run -d -p 8081:8081 shamny/my-api:v2` to check if it worked as expected.

Next, we set up a Minikube cluster locally, which is a single-node Kubernetes cluster, and created Kubernetes manifest files (deployment.yaml and service.yaml).

The Deployment resource in Kubernetes is responsible for managing and maintaining a set of identical Pods. Since Pods are transient and can be created or destroyed, the Deployment ensures that the desired state of the application is always met and available. It also supports features like rolling updates, rollbacks, and scaling.

The Service resource defines a logical set of Pods and a policy for accessing them. Essentially, the Service routes external traffic to the appropriate Pods. In this case, we created a NodePort service, which exposes the service on a fixed port on the node. This allows external access to the service via `nodeip:port`.

Since we are using Minikube, we executed `minikube service golang-svc --url`, which sets up a local proxy to the Minikube cluster and forwards traffic from a port on our local machine to the NodePort (30626) on the Minikube node.

So, when you enter `http://127.0.0.0:53158` in your browser, the request flows through as follows:

Localhost port 53158 → NodePort 30626 → Container port 8081

To access the service, use the URL provided by the `minikube service golang-svc --url` command in your browser to send requests to the API.