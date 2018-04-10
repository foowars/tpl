# tpl

A very simplistic CLI tool that allows rendering arbitrary text/template files,
pulling data in from any YAML file.

There's a docker image:

```
docker pull ripta/tpl
```

Or, get it from the source:

```
go get github.com/ripta/tpl
```

Run it like so:

```
tpl -values=test/data/a.yaml -out=rendered.txt test/templates/ok.tpl
```

Multiple templates can be provided on the command line. Each template is
rendered individually using the same values file, and into the same output
file. The default `-` output file can be used to write to STDOUT.

Multiple value files can be provided as comma-separated paths. Value files are
evaluated in order; later values override earlier ones. For example:

```
tpl -values=test/data/b2.yaml,test/data/b1.yaml -out=rendered.txt test/templates/ok.tpl
```

If `b2.yaml` and `b1.yaml` contain the same keys, then in the above example,
values in `b1.yaml` will override those in `b2.yaml`. If a key exists in
`b2.yaml`, but not in `b1.yaml`, then the value in `b2.yaml` are used.

Optional values may be provided on the command line, but the key and value
would be strings. Command line values override any values that appear in value
files, with the same override rules. For example:

```
tpl -value=foo=bar -value=baz=1234 test/templates/ok.tpl
```

## Nested directories

Nested directory structures are supported. Assuming the following templates:

```
test/templates/deep/ok2.txt.tpl
test/templates/fail.txt.tpl
test/templates/ok.txt.tpl
```

then `tpl ... -out foobar test/templates` will emit:

```
foobar/templates/deep/ok2.txt
foobar/templates/fail.txt
foobar/templates/ok.txt
```

while `tpl ... -out foobar test/templates/*` will emit:

```
foobar/deep/ok2.txt
foobar/fail.txt
foobar/ok.txt
```

## Releasing

```
VERSION=6.0

make test
git tag -a v$VERSION

make build
make push
```
