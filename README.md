# DedicatedContainerIngressController

## Installation

```shell
$ kubectl apply -k manifests
```

## Usage

Applying the following manifest.

```shell
$ cat <<EOS | kubectl apply -f -
apiVersion: ingress.kaidotdev.github.io/v1
kind: DedicatedContainerIngress
metadata:
  name: notebook
spec:
  host: notebook
  template:
    metadata:
      labels:
        app: notebook
    spec:
      containers:
        - name: notebook
          image: jupyter/scipy-notebook:latest
          imagePullPolicy: Always
          command:
            - start-notebook.sh
          args:
            - --NotebookApp.token=''
          ports:
            - name: http
              containerPort: 8888
EOS
```

## How to develop

### `skaffold dev`

```sh
$ make dev
```

### Test

```sh
$ make test
```

### Lint

```sh
$ make lint
```

### Generate CRD from `*_types.go` by controller-gen

```sh
$ make gen
```
