![Archimedes](./ArchimedesLogo.png)
# Archimedes 
## Archimedes is a Kubernetes Operator written to allow for separate repos to control merged configurations using a GitOps approach.

## Artichitecture
![Architecture](./arch.png)

Archimedes is designed to solve configuration issues that become difficult when trying to deploy the same applications to multiple environments where property values differ.   Developers of the applications should not need to worry dymanic platform properties.  They just need to know for their app this is a common set of properties are needed.  Envioroment level properties often change and you don't want developers having to make adjustments to individual application properties every time a platform configuration change is needed.   The solution is to provide a set of properties derived from the Platform repo and at the time of deployment fetch a template from the application repo and merge the 2 toghther to create a configmap that can be configured to be consumed by an application running in the same namespace.  To accomplish this a yaml string is supplied in the spec for the ArchimedesProperty Kubernetes resource type and the information pertaining to the repository where the application code resides containing a template file to be merged.  This template is merged via the go template engine which gives developers the ability to put logic into their tempates if such flexibility is needed.

## ArchimedesProperty

```
apiVersion: backwoods.backwoods-devops.io/v1
kind: ArchimedesProperty
metadata:
  name: archimedesproperty-sample
spec:
  configMapName: 
  RepoUrl:
  Revision;
  PropertiesPath; 
  SourceConfig: |
    env:
      name: test
      databaseConnection: abc123@123
  PropertyType; 
  KeyName: 

## Application Property Template

```
databaseConnection={{ .env.databaseConnection }}