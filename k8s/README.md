Prerequisites

Before starting, make sure your Ubuntu machine has:
    Docker Engine
    kubectl
    k3d
    Helm (optional, for NGINX Ingress or other charts)

Installation comands:

# 1. Install Docker
sudo apt update
sudo apt install -y apt-transport-https ca-certificates curl software-properties-common gnupg lsb-release
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] \
https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io
sudo usermod -aG docker $USER  # log out and log back in

# 2. Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/
kubectl version --client

# 3. Install k3d
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
k3d version

# 4. Install Helm (optional)
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
helm version




# Step 1: Create Kubernetes cluster with k3d

k3d cluster create cloud-storage-cluster \
  --servers 1 \
  --agents 2 \
  --port "80:80@loadbalancer" \
  --port "443:443@loadbalancer"

# Step 2: Apply Kubernetes manifests
kubectl apply -f k8s/

# Step 3: Verify cluster status
kubectl get nodes
kubectl get pods -n cloud-storage
kubectl get svc -n cloud-storage
kubectl get pvc,pv -n cloud-storage
kubectl describe ingress -n cloud-storage

# Step 4: Access the system
Frontend url: http://localhost/
API url: http://localhost/api
MinIO Console: http://localhost:9000