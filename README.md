

  <h2 align="left">K8s namespace quota controller</h3>



---

### About The Project

This is a Kubernetes custom controller.
the Controller itself has been written in GO.

The controller is looking for existing & new namespaces with some custom name in it.

for each namespace, it will try to create some custom ResourceQuota with some custom values from the configFile.

the controller is packaged inside a Docker image.

there is also a deployment for the controller to work inside the cluster.

in order to work it will need some permissions.
so, we first have to create the Service account, the ClusterRole & ClusterRoleBinding


--------



-----------------------------------------



### Provisioning Process

1. Create the SA
  ```sh
  kubectl apply -f manifests/sa.yml
  ```
2. Create the clusterRole
  ```sh
  kubectl apply -f manifests/clusterrole.yml
  ```
3. Create the ClusterRoleBinding
  ```sh
  kubectl apply -f manifests/crb.yml
  ```

4. Create the deployment
  ```sh
  kubectl apply -f ./manifests/deployment.yml
  ```
5. Create new "qa" namespace
  ```sh
  kubectl create namespace qa-dev-env 

  ```
6. Read the pod's logs 
  ```sh
  kubectl logs -n quota-controller quota-controller-7ff8c9bb78-grsjl --follow 

  ```
7. See that a new ResourceQuota object has been applied to the namespace
  ```sh
  Found namespace that contains qa: qa-dev-env 
did not found any resource quotas in namespace qa-dev-env 
creating ResourceQuota in namespace: qa-dev-env 
quota created 1-pod-quota 


  ```
8. Now let's see the behavior of this namespace. apply the test-pod-1.yml
  ```sh
kubectl apply -f test-pod-1.yml -n qa-dev

pod/quota-mem-cpu-demo created
  ```

9. Now that the quota has reached to maximum. let's try to apply the test-pod-2.yml
  ```sh
kubectl apply -f test-pod-2.yml -n qa-dev

Error from server (Forbidden):
 error when creating "test-pod-2.yml": pods "quota-mem-cpu-demo-2" is forbidden:
  exceeded quota: 1-pod-quota, requested: limits.cpu=800m,pods=1, 
  used: limits.cpu=800m,pods=1, limited: limits.cpu=1,pods=1
  ```



----
for more information about this project,

feel free to contact me: 

gaby.tal@tikalk.com

---