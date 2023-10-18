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
  extraMounts:
  - containerPath: /var/lib/kubelet/config.json
    hostPath: "${HOME}/.docker/config.json"
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
if [[ -d ~/.kube ]] && [[ -f ~/.kube/config ]]
then
  export KUBECONFIG=~/.kube/config:/tmp/config
  oc config view --flatten > merged-config.yaml
  mv merged-config.yaml ~/.kube/config
else
  mv /tmp/config ~/.kube/config
fi

oc config use-context kind-kind

# install ingress-nginx
oc apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

# add servicemonitor crd 
oc apply -f ./kind/servicemonitor-crd.yaml

#oc wait --namespace ingress-nginx \
#  --for=condition=ready pod \
#  --selector=app.kubernetes.io/component=controller \
#  --timeout=90s
#
# TODO: add a wait for ingress to be ready with test.yaml & curl
#sleep 20

oc create ns fulcio-system
oc create ns rekor-system
oc -n fulcio-system create secret generic fulcio-secret-rh --from-file=private=./kind/testing-only-cert-key/file_ca_key.pem --from-file=public=./kind/testing-only-cert-key/file_ca_pub.pem --from-file=cert=./kind/testing-only-cert-key/fulcio-root.pem  --from-literal=password=secure --dry-run=client -o yaml | oc apply -f-

oc -n rekor-system create secret generic rekor-private-key --from-file=private=./kind/testing-only-cert-key/rekor_key.pem --dry-run=client -o yaml | oc apply -f-

# install charts
#OPENSHIFT_APPS_SUBDOMAIN=localhost envsubst <  ./examples/values-kind-sigstore.yaml | helm upgrade -i trusted-artifact-signer --debug ./charts/trusted-artifact-signer --wait --wait-for-jobs -n sigstore --create-namespace --values -
