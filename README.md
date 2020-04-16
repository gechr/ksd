# ksd

`ksd` is a quick and handy tool to decode Kubernetes Secrets from standard input (JSON/YAML) and print as YAML.

If [`bat`](https://github.com/sharkdp/bat) is installed, it will automatically be used for syntax highlighting.

## Installation

```sh
$ GO111MODULE=on go get github.com/gechr/ksd@latest
```

## Usage

```
$ kubectl get secret  <name> -o (json|yaml) | ksd
```

## Example

```sh
$ kubectl get secret example -o json
{
    "apiVersion": "v1",
    "data": {
        "abc": "amtsbW5v",
        "def": "cHFyc3R1",
        "ghi": "dnd4eXo="
    },
    "kind": "Secret",
    "metadata": {
        "creationTimestamp": "2019-08-09T08:37:33Z",
        "name": "example",
        "namespace": "default",
        "resourceVersion": "269724870",
        "selfLink": "/api/v1/namespaces/default/secrets/example",
        "uid": "fbb246cc-ace1-44f8-ad19-50db14472ffc"
    },
    "type": "Opaque"
}

$ kubectl get secret example -o json | ksd
apiVersion: v1
kind: Secret
metadata:
  creationTimestamp: "2019-08-09T08:37:33Z"
  name: example
  namespace: default
  resourceVersion: "269724870"
  selfLink: /api/v1/namespaces/default/secrets/example
  uid: 74b4a676-b184-4acc-8cc0-096fe3ca953d
stringData:
  abc: jklmno
  def: pqrstu
  ghi: vwxyz
type: Opaque
```
