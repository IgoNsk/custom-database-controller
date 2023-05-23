# Task

[Task description](task.md)

# Solution

In controller loop first step - recieve actual inforamtion about CustmoDatabse CRD from k8s - we didn't use local storage 
for this information.

If CRD object found - we choose action AddOrUpdate. First step - create all object in Postgresql. We try to create database, user and rant user to database. If there
are existed database and not exists Secret - it means than Secret Name was changed or previous action of 
Creation was interrupted. And we continue process of creation and change user password. 
Second step - if all Postgresql object was created successful - we create new Secret with DB creds. 
If Secret already exists - we wouldn't change it, because we have no things, that may update in Secret.

If CRD object not found - we choose action Delete. It means, that we are deleting created user and database in Postgresql.
Secret will be deleted automatically by k8s, because we use Owner section and linkin with CRD in creation of Secret.

To manager Postgresql databases and users i choose SQL interface and commands. This way incapsulated in separated component.

## Application Design

I separate the application by some layers:

* **cmd/customdatabase-controller** - it is entrypoint of our application. On this layer we create instance of application controller 
and mix all needed resources together manually, without any DI frameworks. We use CLI flags to configure our application.
* **pkg/*** - it's a place for common libraries, code than could be shared between some parts of our application.
* **internal/customdatabase** - there are all parts of our internal logic application. It's very simple implementation 
of Clean Architecture. I choose only important ideas from this conception, and it's not Clean by 100% :)
* * **in root** of package there pure domain logic service and domain entities. In our case - 
entity about created CustomDatabase and service that reproduce this entity. On this layer we have no any details 
of external dependencies - storages, queues.
* * **usecases** - there are logic, that represent interaction with real world. In our case - it's a controller of CRD resource
with specific action handlers. On this layer we know haw to communicate with k8s and his resources. We also describe
our business scenarios and represent they in handler for specific actions. We also use domain logics and other external 
dependencies by interface. Tests on this layer is pure unit test and no need any external dependencies with real instances (db, k8s)
* * **adapters** this layer for implementation of service? that uses in UseCase layer by interface. On this layer we 
know about specific details of used technology, like postgresql driver.

## How to run

**Important!** You should have installed Minikube on you local environment and golang version 1.19 or higher

```bash
    # this step initialize application locally: prepare and build controller application
    # register CRD in local k8s, and , of course, run application.
    make dev-run
    
    # this step demonstrates basics use cases of CustomDatabases
    make dev-integration-test
    
    # to run unit test in project
    make test
```

## TODO

* Add validation in CRD schema for secretName field
* Check and prevent collision in using one secret. More than one CustomDatabases may use the same SecretName, but we use one secret only for one DB
* Add ability to generate random password for created CustomDatabase users
* Add tests for error cases
* Add linters for project