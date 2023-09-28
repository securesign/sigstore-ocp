# run this from root of repository

# spin up kind cluster
cat <<EOF | sudo kind create cluster --image kindest/node:v1.28.0 --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF

sudo kind get kubeconfig > /tmp/config
sudo chown $USER:$USER /tmp/config
mv /tmp/config ~/.kube/config

# install ingress-nginx
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

#kubectl wait --namespace ingress-nginx \
#  --for=condition=ready pod \
#  --selector=app.kubernetes.io/component=controller \
#  --timeout=90s
#
# TODO: add a wait for ingress to be ready with test.yaml & curl
#sleep 20

oc create ns fulcio-system
oc create ns rekor-system
oc -n fulcio-system create secret generic fulcio-secret-rh --from-file=private=./kind/test-keys-cert/file_ca_key.pem --from-file=public=./kind/test-keys-cert/file_ca_pub.pem --from-file=cert=./kind/test-keys-cert/fulcio-root.pem  --from-literal=password=secure --dry-run=client -o yaml | oc apply -f-

oc -n rekor-system create secret generic rekor-private-key --from-file=private=./kind/test-keys-cert/rekor_key.pem --dry-run=client -o yaml | oc apply -f-

# install charts
#OPENSHIFT_APPS_SUBDOMAIN=localhost envsubst <  ./examples/values-kind-sigstore.yaml | helm upgrade -i trusted-artifact-signer --debug ./charts/trusted-artifact-signer -n sigstore --create-namespace --values -
