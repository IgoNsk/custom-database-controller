# Kubernetes shared database controller

## The Problem
Many users want to test their web services or make a simple service for help with actual development. Most of the time web apps require databases like Postgres or Mysql. But managed databases in Clouds are expensive and usually require special permissions and time for deployment. Also, most of the time the query load is very low for applications.

## The Goal

The goal is to help users in Kubernetes to get fast(deployment), reliable service which will create db(schemas) and users at the one HA instance of the database per Kubernetes Cluster.

Good idea is to create a [Kubernetes Controller](https://kubernetes.io/docs/concepts/architecture/controller/). It will watch for the database objects like this:

```yaml
apiVersion: v1
kind: CustomDatabase
metadata:
  name: name
  namespace: namespace
spec:
  secretName: secretName
```

And when a user creates an object, it will create a database and user by the metadata.name and put the generated password, host, port, db_name etc to the Kubernetes Secret with name secretName. When a user deletes an object, the database and db user will be deleted with all data(w/o confirmation etc. due to scope of the task). Update of the object is out of scope(changing secret name, changing referenced secret etc)

You can take Minikube and pod with postgres as a test stand.


You can ask directly in email any questions and we will be happy to answer and help.

Sample controller: https://github.com/kubernetes/sample-controller
