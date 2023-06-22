#!/usr/bin/env sh

oc adm policy add-scc-to-user anyuid -z fulcio-createcerts -n fulcio-system
oc adm policy add-scc-to-user anyuid -z fulcio-server -n fulcio-system
oc adm policy add-scc-to-user anyuid -z ctlog -n ctlog-system
oc adm policy add-scc-to-user anyuid -z ctlog-createtree -n ctlog-system
oc adm policy add-scc-to-user anyuid -z scaffolding-ctlog-createctconfig -n ctlog-system
oc adm policy add-scc-to-user anyuid -z tuf -n tuf-system
oc adm policy add-scc-to-user anyuid -z tuf-secret-copy-job -n tuf-system
oc adm policy add-scc-to-user anyuid -z trillian-logserver -n trillian-system
oc adm policy add-scc-to-user anyuid -z trillian-mysql -n trillian-system
oc adm policy add-scc-to-user anyuid -z trillian-logsigner -n trillian-system
oc adm policy add-scc-to-user anyuid -z rekor-redis -n rekor-system
oc adm policy add-scc-to-user anyuid -z scaffolding-rekor-createtree -n rekor-system
oc adm policy add-scc-to-user anyuid -z rekor-server -n rekor-system
