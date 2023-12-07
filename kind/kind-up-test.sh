# run this from root of repository

# spin up kind cluster
cat <<EOF | kind create cluster --image kindest/node:v1.28.0 --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraMounts:
  - containerPath: /var/lib/kubelet/config.json
    hostPath: "${HOME}/.docker/config.json"
EOF

kind get kubeconfig > /tmp/config
chown $USER:$USER /tmp/config
if [[ -d ~/.kube ]] && [[ -f ~/.kube/config ]]
then
  export KUBECONFIG=~/.kube/config:/tmp/config
  oc config view --flatten > merged-config.yaml
  mv merged-config.yaml ~/.kube/config
else
  mv /tmp/config ~/.kube/config
fi
chmod go-r ~/.kube/config

oc config use-context kind-kind

oc create ns fulcio-system
oc create ns rekor-system
oc -n fulcio-system create secret generic fulcio-secret-rh --from-file=private=./kind/testing-only-cert-key/file_ca_key.pem --from-file=public=./kind/testing-only-cert-key/file_ca_pub.pem --from-file=cert=./kind/testing-only-cert-key/fulcio-root.pem  --from-literal=password=secure --dry-run=client -o yaml | oc apply -f-

oc -n rekor-system create secret generic rekor-private-key --from-file=private=./kind/testing-only-cert-key/rekor_key.pem --dry-run=client -o yaml | oc apply -f-

# install charts
helm upgrade -i trusted-artifact-signer --debug ./charts/trusted-artifact-signer --wait --wait-for-jobs --timeout 10m -n trusted-artifact-signer --create-namespace --values ./examples/values-kind-sigstore.yaml && \
helm test trusted-artifact-signer -n trusted-artifact-signer
