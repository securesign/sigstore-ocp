# Updating the pull secret

### Why do it

When your Openshift cluster being provisioned, it uses a pull secret to identify you against many of the services offered by redhat such as Quay, Openshift, Registry.Redhat.io, etc, which is saved to the cluster after installation. To properly install the `trusted-artifact-signer` stack we have to add an entry for `registry.redhat.io`, to ensure who is using the service, and that they have the proper permissions for their ogranization. For any additional questions on this topic, refer to the [Openshift Documentation on using pull-secrets](https://docs.openshift.com/container-platform/4.13/openshift_images/managing_images/using-image-pull-secrets.html#images-update-global-pull-secret_using-image-pull-secrets).

## Steps

Firstly download the pull-secret data:

```bash
oc get secret/pull-secret -n openshift-config --template='{{index .data ".dockerconfigjson" | base64decode}}' > ./pull-secret
```

Add new credentials to the pull-secret data:

```bash
oc registry login --registry="registry.redhat.io" \ 
--auth-basic="<username>:<password>" \ 
--to=./pull-secret
```

Finally, update the pull-secret:

```bash
oc set data secret/pull-secret -n openshift-config --from-file=.dockerconfigjson=./pull-secret
```



