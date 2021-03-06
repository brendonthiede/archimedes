![Archimedes](./ArchimedesLogo.png)

# Archimedes

## Archimedes is a Kubernetes Operator written to allow for separate repos to control merged configurations for different environments using a GitOps approach.

## Architecture

![Architecture](./arch.png)

Archimedes is designed to simplify configuration complexity when trying to deploy the same applications to multiple environments where property values differ.  Developers of the applications should not need to worry about dynamic platform properties.  They just need to know for their app, what properties are needed for the application to run.  Environment level properties often change and you don't want developers having to make adjustments to individual application properties every time a platform configuration change is needed.   The solution is to provide a set of properties derived from the Platform repo and at the time of deployment fetch a template from the application repo and merge the 2 together to create a configmap that can be configured to be consumed by an application running in the same namespace.  To accomplish this a yaml string is supplied in the sourceConfig property in the ArchimedesProperty Kubernetes resource type specification along with the information pertaining to the repository where the application code resides containing a template file to be merged.  This template is merged via the go template engine which gives developers the ability to put logic into their templates if such flexibility is needed.

## Installation

A helm chart for deployment is supplied in the chart/archimedes-property-operator directory.  This chart also includes the archimedes custom resource definition.

```sh
cd chart/archimedes-property-operator
```

```sh
kubectl apply -f crd/archimedes.backwoods-devops.io_archimedesproperties.yaml
```

Next steps would be to setup needed configuration and supply the credentials for your repo including a CA certificate if needed. Once that is ready you can install the operator.

```sh
helm install archimedes-property-operator .
```

## Usage

To deploy an Archimedes property you will need an application property template along with a Archimedes property  definition.  The properties template file will be stored in with the application repo or any repo desired.  This template will be parsed using the [golang text/template package](https://pkg.go.dev/text/template "text/template package").  A key value pair pattern is used with the `=` to separate the key and the value to define the individual properties.   Data to be merged is supplied from the sourceConfig in the ArchimedesProperty definition. There are two options for how this template is parsed into the configMap that will be generated.  The propertyType in the Archimedes property spec (details below) allows for either individual key values to be parsed into separate values (kvp option and values in template separated by `=`) or the final merged template to be setup under a defined key value (file option with keyName property defined).  The file option opens the door for any configuration format to be merged with the sourceConfig.

### Application Property Template

```ini
databaseName={{ .env.dbanme }}
databasePort={{ .env.dbport }}
{{- if ne .env.name "Production"}}
templateCache=true
{{- else}}
templateCache=false
{{- end}}
```

Next thing needed is the ArchimedesProperty definition.

### Spec:

| Name | Description | Type |
| ----- | ----------- | ------- |
| name | name of the configmap to be created | string |
| repoURL | url to the repo containing the property template to be merged | string |
| revision | the commit hash, branch or tag | string |
| caPath | path to CA Certificate for git repo to use if required | string |
| propertiesPath | the path to the template file | string |
| sourceConfig | a yaml configuration file supplied by the platform/env | string |
| propertyType | configmap data style.  Options are kvp or key.  kvp will create a separate entry for each line in the properties template (values in kvp values in template file are separated by `=` ).   key will place the results of the merged template as a string value under the name defined in keyName. Use this method if you have a configuration to be consumed that is not in a kvp format. | string |
| keyName | name of the key template results are saved to.  Only applies when propertyType is set to key | string |

### ArchimedesProperty

```yaml
apiVersion: archimedes.backwoods-devops.io/v1
kind: ArchimedesProperty
metadata:
  name: archimedesproperty-trees-app
  namespace: default
spec:
  configMapName: trees-app-properties
  repoUrl: "https://github.com/backwoods-devops/archimedes.git"
  revision: main
  propertiesPath: config/samples/properties.tpl
  sourceConfig: |
    env:
      name: staging
      dbname: forest-data
      dbport: 5432
  propertyType: key
  keyName: config.properties
```

### Deploy your property

```sh
cat <<EOF | kubectl apply -f -
apiVersion: archimedes.backwoods-devops.io/v1
kind: ArchimedesProperty
metadata:
  name: archimedesproperty-trees-app
  namespace: default
spec:
  configMapName: trees-app-properties
  repoUrl: "https://github.com/backwoods-devops/archimedes.git"
  revision: main
  propertiesPath: config/samples/properties.tpl
  sourceConfig: |
    env:
      name: staging
      dbname: forest-data
      dbport: 5432
  propertyType: key
  keyName: config.properties
EOF
```

## Extra properties added

There will be several properties automatically added.

commit, repoUrl, revision and path will be populated so they may be referenced as needed by your tooling to determine proper versioning

## Handy tips

This project was built using kubebuilder.   Please visit the [Kubebuilder book](https://book.kubebuilder.io/ "Kubebuilder Book") website for more info on building this project.

Leverage a gitops tool such as [Argo CD](https://argoproj.github.io/cd/ "Argo CD") or [Flux](https://fluxcd.io/ "Flux").  Using these tools pass the desired sourceConfig or use helm chart to take multiple values and build out the sourceConfig yaml as desired for each env and deploy the AchimediesProperty file using that.
