helm install postgres oci://registry-1.docker.io/bitnamicharts/postgresql -f deployments/postgres/values.yaml --wait
helm install kafka oci://registry-1.docker.io/bitnamicharts/kafka -f deployments/kafka/values.yaml --wait
helm repo add gruntwork https://helmcharts.gruntwork.io
helm repo update
helm install cars-api gruntwork/k8s-service -f deployments/api-services/cars.yaml --wait
helm install rental-api gruntwork/k8s-service -f deployments/api-services/rental.yaml --wait
helm install payment-api gruntwork/k8s-service -f deployments/api-services/payment.yaml --wait
# kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/cloud/deploy.yaml
helm install gateway gruntwork/k8s-service -f deployments/api-services/gateway.yaml --wait
