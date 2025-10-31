# Docs

Welcome to the documentation.</br> 
The code and chart is avalible at [github.com](https://github.com/callmewhatuwant/sealed-age-operator).

## Getting started

* install

```bash
helm repo add sealed-age-operator \
https://callmewhatuwant.github.io/sealed-age-operator
helm install sealed-age-operator sealed-age-operator/age-secrets \ 
--namespace sealed-age-system --create-namespace
```

* check install

```bash
kubectl wait --for=condition=Ready pods --all -n sealed-age-system 
```

* uninstall

```bash
helm uninstall -n sealed-age-system age-secrets
kubectl delete namespace sealed-age-system
```

## First secret

* install age

```bash
sudo apt install age
```

* get key

```bash
LATEST=$(kubectl get secrets -n sealed-age-system --no-headers -o custom-columns=":metadata.name" \
  | grep '^age-key-' | sort | tail -n1)

kubectl get secret "$LATEST" -n sealed-age-system -o jsonpath='{.data.public}' | base64 --decode && echo
```

* create test file

```bash
echo test123 > secret.txt
```

* encrypt with ur public key

```bash
age --armor -r age1u4dtwstnutaytrfjea9jp3v9y0a8l9hh7rlgmehz9w63z0u3zuvquxhhhy secret.txt
```

* create crd ressource 
* crd has to be applied before doing that

```bash
kubectl apply -f kubectl apply -f https://raw.githubusercontent.com/callmewhatuwant/sealed-age-operator/main/config/crd/bases/security.age.io_sealedages.yaml
```

* exmaple secret crd ressource

```yaml
apiVersion: security.age.io/v1alpha1
kind: SealedAge
metadata:
  name: db-passwd
spec:
  encryptedData:
    password: |
      -----BEGIN AGE ENCRYPTED FILE-----
      YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSBWbHhqcGhyZ0ZSbXhQZXJ1
      aU1kL1NmZjYyaU9JQXlQazBuekdmMk8ySkYwCkloMGJxR0lXVG0yM2FXV3hrT3BI
      OXVwdzhrYWtGU0hwTUtLTHN5dzJBTGsKLS0tIEc0V1JmTUVpWkZuNGFGWXJJV3ow
      cWZpL09JTnFCVFFZbXRFQUY2QTdTbm8KdkZOvCXRqENpCw9ncrVP+qzDBTKwntfi
      ihgfMGuoy3Q37Dkqsw==
      -----END AGE ENCRYPTED FILE-----
```

* verify

```bash
kubectl get secret -n sealed-age-system
```

## Helm Options

```yaml
## name override
fullnameOverride: sealed-age-controller
sealedAgeController:

## leader election
  leaderElection:
    enabled: true
    namespace: sealed-age-system

  ## replicas for ha
  replicas: 3

  controller:
    ## image
    image:
      repository: callmewhatuwant/age-secrets-operator
      tag: 0.0.2
    imagePullPolicy: IfNotPresent
    
    ## resources
    resources:
      limits:
        cpu: 200m
        memory: 128Mi
      requests:
        cpu: 100m
        memory: 64Mi

    ## security
    containerSecurityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL
      runAsNonRoot: true
      runAsUser: 65532

## prometheus
metricsService:
  type: ClusterIP
  ports:
    - port: 8080
      name: metrics
      targetPort: 8080

## monitor for prometheus
ServiceMonitor:
  enabled: true
  endpoints:
    - port: metrics
      interval: 30s 
      path: /metrics

## job
ageKeyRotation:
  schedule: "0 0 1 * *"
  
  ## initial key
  initialRun:
    enabled: true

  ## image for cron and init job
  image:
    repository: alpine
    tag: "3.20"
    pullPolicy: IfNotPresent
```

## Support me if you want

BTC:

```bash
bc1q7zgprykqzj4vprzxzafy5lskhpv7qau9p7a28r
```

Solana:
```bash
B6aGswkR4tpYDCaLny4B1rZWwQNrDk4dEvpEGjJw3GGG
```

<div style="text-align:center;">
  <img src="images/qr-btc.png" alt="BTC QR" width="300" height="300" style="display:inline-block; margin-right:30px;">
  <img src="images/qr-sol.png" alt="SOL QR" width="300" height="300" style="display:inline-block;">
</div>
