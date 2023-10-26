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

oc config use-context kind-kind

oc create ns fulcio-system
oc create ns rekor-system
oc -n fulcio-system create secret generic fulcio-secret-rh --from-file=private=./kind/testing-only-cert-key/file_ca_key.pem --from-file=public=./kind/testing-only-cert-key/file_ca_pub.pem --from-file=cert=./kind/testing-only-cert-key/fulcio-root.pem  --from-literal=password=secure --dry-run=client -o yaml | oc apply -f-

oc -n rekor-system create secret generic rekor-private-key --from-file=private=./kind/testing-only-cert-key/rekor_key.pem --dry-run=client -o yaml | oc apply -f-

#install OLM
kubectl create -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.25.0/crds.yaml
# wait for a while to be sure CRDs are installed
sleep 1
kubectl create -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.25.0/olm.yaml

#install keycloak from Kind overlay
kubectl create --kustomize keycloak/operator/overlay/kind
until [ ! -z "$(kubectl get pod -l name=keycloak-operator -n keycloak-system 2>/dev/null)" ]
do
  echo "Waiting for keycloak operator. Pods in keycloak-system namespace:"
  kubectl get pods -n keycloak-system
  sleep 10
done
kubectl create --kustomize keycloak/resources/overlay/kind
until [[ $( oc get keycloak keycloak -o jsonpath='{.status.ready}' -n keycloak-system 2>/dev/null) == "true" ]]
do
  printf "Waiting for keycloak deployment. \n Keycloak ready: %s \n" $(oc get keycloak keycloak -o jsonpath='{.status.ready}' -n keycloak-system)
  sleep 10
done

# install charts
helm upgrade -i trusted-artifact-signer --debug ./charts/trusted-artifact-signer --wait --wait-for-jobs --timeout 10m -n sigstore --create-namespace --values ./examples/values-kind-sigstore.yaml && \
helm test trusted-artifact-signer -n sigstore
