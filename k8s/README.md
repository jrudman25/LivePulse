# LivePulse backend on local K3s

This local stack runs the Go backend, PostgreSQL, and Redis in the `k3d-livepulse-lab` cluster. It does not connect to the production Neon or Upstash endpoints in `backend/.env`. Clerk remains an external dependency using the existing test-mode secret. Ticketmaster is disabled on the first run.

## Local data and credentials

The backend creates its schema in the in-cluster PostgreSQL database during startup. PostgreSQL and Redis each use a 1-node StatefulSet and a local-path PVC. The volumes survive pod restarts, but are local to this k3d cluster; deleting the namespace or cluster can delete the lab data.

The secret helper generates random local database passwords, preserves them on reruns, and reads only the optional Clerk, admin, and explicitly enabled Ticketmaster keys from `backend/.env`. It ignores `DATABASE_URL` and `REDIS_URL`, so the production endpoints are not copied into Kubernetes. The Secret is stored in the local cluster and is not a substitute for an encrypted production secrets store.

Ticketmaster stays disabled unless you explicitly pass `-EnableTicketmaster`. Enabling it causes an immediate fetch when the backend starts, then another every six hours; fetched events are written to local PostgreSQL and expired events are deleted.

## Build and deploy

Run these commands from the repository root in PowerShell. Start Docker Desktop first. Create the cluster once if it does not exist yet; for an existing stopped cluster, use `k3d cluster start livepulse-lab` instead.

```powershell
# Run this only if the cluster does not already exist.
k3d cluster create livepulse-lab
kubectl config use-context k3d-livepulse-lab
docker build -t livepulse-backend:dev -f backend/Dockerfile backend
k3d image import livepulse-backend:dev --cluster livepulse-lab
.\k8s\create-local-secret.ps1
kubectl apply -f .\k8s\local-services.yaml
kubectl rollout status statefulset/livepulse-postgres --namespace livepulse --timeout=180s
kubectl rollout status statefulset/livepulse-redis --namespace livepulse --timeout=180s
kubectl apply -f .\k8s\backend.yaml
kubectl rollout status deployment/livepulse-backend --namespace livepulse --timeout=180s
kubectl get pods,services,pvc --namespace livepulse
kubectl logs deployment/livepulse-backend --namespace livepulse --tail=50
```

The local-path provisioner bundled with K3s supplies the database volumes. The backend is a `ClusterIP` Service, so reach it through a port-forward:

```powershell
kubectl port-forward service/livepulse-backend 18081:8080 --namespace livepulse
```

In another terminal, check the process and dependency readiness:

```powershell
curl.exe -i http://127.0.0.1:18081/health
curl.exe -i http://127.0.0.1:18081/health/ready
```

`/health` checks the process. `/health/ready` pings the in-cluster PostgreSQL and Redis services. CORS and WebSocket origins allow a local frontend on ports 3000.
